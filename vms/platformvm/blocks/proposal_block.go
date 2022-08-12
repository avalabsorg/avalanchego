// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ Block = &BlueberryProposalBlock{}
	_ Block = &ApricotProposalBlock{}
)

func NewBlueberryProposalBlock(
	timestamp time.Time,
	parentID ids.ID,
	height uint64,
	tx *txs.Tx,
) (Block, error) {
	res := &BlueberryProposalBlock{
		BlueberryCommonBlock: BlueberryCommonBlock{
			ApricotCommonBlock: ApricotCommonBlock{
				PrntID: parentID,
				Hght:   height,
			},
			BlkTimestamp: uint64(timestamp.Unix()),
		},
		Tx: tx,
	}

	return res, initialize(Block(res))
}

type BlueberryProposalBlock struct {
	BlueberryCommonBlock `serialize:"true"`

	Tx *txs.Tx `serialize:"true" json:"tx"`
}

func (b *BlueberryProposalBlock) initialize(bytes []byte) error {
	b.BlueberryCommonBlock.initialize(bytes)
	if err := b.Tx.Sign(txs.Codec, nil); err != nil {
		return fmt.Errorf("failed to initialize tx: %w", err)
	}
	return nil
}

func (b *BlueberryProposalBlock) Txs() []*txs.Tx { return []*txs.Tx{b.Tx} }

func (b *BlueberryProposalBlock) Visit(v Visitor) error {
	return v.BlueberryProposalBlock(b)
}

func NewApricotProposalBlock(
	parentID ids.ID,
	height uint64,
	tx *txs.Tx,
) (Block, error) {
	res := &ApricotProposalBlock{
		ApricotCommonBlock: ApricotCommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
		Tx: tx,
	}

	return res, initialize(Block(res))
}

// As is, this is duplication of atomic block. But let's tolerate some code duplication for now
type ApricotProposalBlock struct {
	ApricotCommonBlock `serialize:"true"`

	Tx *txs.Tx `serialize:"true" json:"tx"`
}

func (b *ApricotProposalBlock) initialize(bytes []byte) error {
	b.ApricotCommonBlock.initialize(bytes)
	if err := b.Tx.Sign(txs.Codec, nil); err != nil {
		return fmt.Errorf("failed to initialize tx: %w", err)
	}
	return nil
}

func (b *ApricotProposalBlock) Txs() []*txs.Tx { return []*txs.Tx{b.Tx} }

func (b *ApricotProposalBlock) Visit(v Visitor) error {
	return v.ApricotProposalBlock(b)
}
