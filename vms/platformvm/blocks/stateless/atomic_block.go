// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateless

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var _ AtomicBlockIntf = &AtomicBlock{}

type AtomicBlockIntf interface {
	CommonBlockIntf

	// ProposalTx returns list of transactions
	// contained in the block
	AtomicTx() *txs.Tx
}

func NewAtomicBlock(parentID ids.ID, height uint64, tx *txs.Tx) (AtomicBlockIntf, error) {
	res := &AtomicBlock{
		CommonBlock: CommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
		Tx: tx,
	}

	// We serialize this block as a Block so that it can be deserialized into a
	// Block
	blk := CommonBlockIntf(res)
	bytes, err := Codec.Marshal(ApricotVersion, &blk)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
	}

	if err := tx.Sign(txs.Codec, nil); err != nil {
		return nil, fmt.Errorf("failed to sign block: %w", err)
	}

	return res, res.Initialize(ApricotVersion, bytes)
}

// AtomicBlock being accepted results in the atomic transaction contained in the
// block to be accepted and committed to the chain.
type AtomicBlock struct {
	CommonBlock `serialize:"true"`

	Tx *txs.Tx `serialize:"true" json:"tx"`
}

func (ab *AtomicBlock) Initialize(version uint16, bytes []byte) error {
	if err := ab.CommonBlock.Initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	unsignedBytes, err := txs.Codec.Marshal(txs.Version, &ab.Tx.Unsigned)
	if err != nil {
		return fmt.Errorf("failed to marshal unsigned tx: %w", err)
	}
	signedBytes, err := txs.Codec.Marshal(txs.Version, &ab.Tx)
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %w", err)
	}
	ab.Tx.Initialize(unsignedBytes, signedBytes)
	return nil
}

func (ab *AtomicBlock) AtomicTx() *txs.Tx { return ab.Tx }
