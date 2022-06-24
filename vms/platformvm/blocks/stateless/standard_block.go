// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateless

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ StandardBlockIntf = &StandardBlock{}
	_ StandardBlockIntf = &PostForkStandardBlock{}
)

type StandardBlockIntf interface {
	CommonBlockIntf

	// DecisionTxs returns list of transactions
	// contained in the block
	DecisionTxs() []*txs.Tx
}

func NewStandardBlock(
	version uint16,
	timestamp uint64,
	parentID ids.ID,
	height uint64,
	txes []*txs.Tx,
) (StandardBlockIntf, error) {
	// make sure txs to be included in the block
	// are duly initialized
	for _, tx := range txes {
		if err := tx.Sign(txs.Codec, nil); err != nil {
			return nil, fmt.Errorf("failed to sign block: %w", err)
		}
	}

	switch version {
	case PreForkVersion:
		res := &StandardBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			Txs: txes,
		}
		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := CommonBlockIntf(res)
		bytes, err := Codec.Marshal(version, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}

		return res, res.Initialize(version, bytes)

	case PostForkVersion:
		txsBytes := make([][]byte, 0, len(txes))
		for _, tx := range txes {
			txsBytes = append(txsBytes, tx.Bytes())
		}
		res := &PostForkStandardBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			TxsBytes: txsBytes,
			Txs:      txes,
		}
		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := CommonBlockIntf(res)
		bytes, err := Codec.Marshal(version, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}
		return res, res.Initialize(version, bytes)

	default:
		return nil, fmt.Errorf("unsopported block version %d", version)
	}
}

type StandardBlock struct {
	CommonBlock `serialize:"true"`

	Txs []*txs.Tx `serialize:"true" json:"txs"`
}

func (sb *StandardBlock) Initialize(version uint16, bytes []byte) error {
	if err := sb.CommonBlock.Initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	for _, tx := range sb.Txs {
		if err := tx.Sign(txs.Codec, nil); err != nil {
			return fmt.Errorf("failed to sign block: %w", err)
		}
	}
	return nil
}

func (sb *StandardBlock) DecisionTxs() []*txs.Tx { return sb.Txs }

type PostForkStandardBlock struct {
	CommonBlock `serialize:"true"`

	TxsBytes [][]byte `serialize:"false" postFork:"true" json:"txs"`

	Txs []*txs.Tx
}

func (psb *PostForkStandardBlock) Initialize(version uint16, bytes []byte) error {
	if err := psb.CommonBlock.Initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	txes := make([]*txs.Tx, 0, len(psb.TxsBytes))
	for _, txBytes := range psb.TxsBytes {
		var tx txs.Tx
		_, err := txs.Codec.Unmarshal(txBytes, &tx)
		if err != nil {
			return fmt.Errorf("failed unmarshalling tx in post fork block: %w", err)
		}
		if err := tx.Sign(txs.Codec, nil); err != nil {
			return fmt.Errorf("failed to sign block: %w", err)
		}
		txes = append(txes, &tx)
	}
	psb.Txs = txes
	return nil
}

func (psb *PostForkStandardBlock) DecisionTxs() []*txs.Tx { return psb.Txs }
