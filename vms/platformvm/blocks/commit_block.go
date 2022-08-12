// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ Block = &BlueberryCommitBlock{}
	_ Block = &ApricotCommitBlock{}
)

func NewBlueberryCommitBlock(timestamp time.Time, parentID ids.ID, height uint64) (Block, error) {
	res := &BlueberryCommitBlock{
		BlkTimestamp: uint64(timestamp.Unix()),
		ApricotCommitBlock: &ApricotCommitBlock{
			ApricotCommonBlock: ApricotCommonBlock{
				PrntID: parentID,
				Hght:   height,
			},
		},
	}

	return res, initialize(Block(res))
}

type BlueberryCommitBlock struct {
	BlkTimestamp uint64 `serialize:"true" json:"time"`

	*ApricotCommitBlock `serialize:"true"`
}

func (b *BlueberryCommitBlock) BlockTimestamp() time.Time {
	return time.Unix(int64(b.BlkTimestamp), 0)
}

func (b *BlueberryCommitBlock) Visit(v Visitor) error {
	return v.BlueberryCommitBlock(b)
}

func NewApricotCommitBlock(parentID ids.ID, height uint64) (Block, error) {
	res := &ApricotCommitBlock{
		ApricotCommonBlock: ApricotCommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
	}

	return res, initialize(Block(res))
}

type ApricotCommitBlock struct {
	ApricotCommonBlock `serialize:"true"`
}

func (b *ApricotCommitBlock) initialize(bytes []byte) error {
	b.ApricotCommonBlock.initialize(bytes)
	return nil
}

func (*ApricotCommitBlock) Txs() []*txs.Tx { return nil }

func (b *ApricotCommitBlock) Visit(v Visitor) error {
	return v.ApricotCommitBlock(b)
}
