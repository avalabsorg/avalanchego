// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p

import (
	"fmt"
	"sync"

	stdcontext "context"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
)

var _ Backend = &backend{}

type ChainUTXOs interface {
	AddUTXO(ctx stdcontext.Context, destinationChainID ids.ID, utxo *avax.UTXO) error
	RemoveUTXO(ctx stdcontext.Context, sourceChainID, utxoID ids.ID) error

	UTXOs(ctx stdcontext.Context, sourceChainID ids.ID) ([]*avax.UTXO, error)
	GetUTXO(ctx stdcontext.Context, sourceChainID, utxoID ids.ID) (*avax.UTXO, error)
}

// Backend defines the full interface required to support a P-chain wallet.
type Backend interface {
	ChainUTXOs
	BuilderBackend
	SignerBackend

	AcceptTx(ctx stdcontext.Context, tx *signed.Tx) error
}

type backend struct {
	Context
	ChainUTXOs

	txsLock sync.RWMutex
	// txID -> tx
	txs map[ids.ID]*signed.Tx
}

func NewBackend(ctx Context, utxos ChainUTXOs, txs map[ids.ID]*signed.Tx) Backend {
	return &backend{
		Context:    ctx,
		ChainUTXOs: utxos,
		txs:        txs,
	}
}

func (b *backend) AcceptTx(ctx stdcontext.Context, tx *signed.Tx) error {
	var baseTx *unsigned.BaseTx
	txID := tx.Unsigned.ID()
	switch utx := tx.Unsigned.(type) {
	case *unsigned.AddDelegatorTx:
		baseTx = &utx.BaseTx
	case *unsigned.AddSubnetValidatorTx:
		baseTx = &utx.BaseTx
	case *unsigned.AddValidatorTx:
		baseTx = &utx.BaseTx
	case *unsigned.ExportTx:
		baseTx = &utx.BaseTx

		for i, out := range utx.ExportedOutputs {
			err := b.AddUTXO(
				ctx,
				utx.DestinationChain,
				&avax.UTXO{
					UTXOID: avax.UTXOID{
						TxID:        txID,
						OutputIndex: uint32(len(utx.Outs) + i),
					},
					Asset: avax.Asset{ID: out.AssetID()},
					Out:   out.Out,
				},
			)
			if err != nil {
				return err
			}
		}
	case *unsigned.ImportTx:
		baseTx = &utx.BaseTx

		consumedRemoteUTXOIDs := utx.InputUTXOs()
		err := b.removeUTXOs(ctx, utx.SourceChain, consumedRemoteUTXOIDs)
		if err != nil {
			return err
		}
	case *unsigned.CreateChainTx:
		baseTx = &utx.BaseTx
	case *unsigned.CreateSubnetTx:
		baseTx = &utx.BaseTx
	default:
		return fmt.Errorf("%w: %T", errUnknownTxType, tx.Unsigned)
	}

	consumedUTXOIDs := baseTx.InputIDs()
	err := b.removeUTXOs(ctx, constants.PlatformChainID, consumedUTXOIDs)
	if err != nil {
		return err
	}

	producedUTXOSlice := baseTx.UTXOs()
	err = b.addUTXOs(ctx, constants.PlatformChainID, producedUTXOSlice)
	if err != nil {
		return err
	}

	b.txsLock.Lock()
	defer b.txsLock.Unlock()

	b.txs[txID] = tx
	return nil
}

func (b *backend) addUTXOs(ctx stdcontext.Context, destinationChainID ids.ID, utxos []*avax.UTXO) error {
	for _, utxo := range utxos {
		if err := b.AddUTXO(ctx, destinationChainID, utxo); err != nil {
			return err
		}
	}
	return nil
}

func (b *backend) removeUTXOs(ctx stdcontext.Context, sourceChain ids.ID, utxoIDs ids.Set) error {
	for utxoID := range utxoIDs {
		if err := b.RemoveUTXO(ctx, sourceChain, utxoID); err != nil {
			return err
		}
	}
	return nil
}

func (b *backend) GetTx(_ stdcontext.Context, txID ids.ID) (*signed.Tx, error) {
	b.txsLock.RLock()
	defer b.txsLock.RUnlock()

	tx, exists := b.txs[txID]
	if !exists {
		return nil, database.ErrNotFound
	}
	return tx, nil
}
