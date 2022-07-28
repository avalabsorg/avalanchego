// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateless

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ Block = &AbortBlock{}
	_ Block = &CommitBlock{}
)

type AbortBlock struct {
	CommonBlock `serialize:"true"`
}

func (ab *AbortBlock) BlockTxs() []*txs.Tx { return nil }

func (ab *AbortBlock) Visit(v Visitor) error {
	return v.AbortBlock(ab)
}

func NewAbortBlock(
	parentID ids.ID,
	height uint64,
) (*AbortBlock, error) {
	res := &AbortBlock{
		CommonBlock: CommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
	}

	// We serialize this block as a Block so that it can be deserialized into a
	// Block
	blk := Block(res)
	bytes, err := Codec.Marshal(txs.Version, &blk)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
	}

	return res, res.initialize(bytes)
}

type CommitBlock struct {
	CommonBlock `serialize:"true"`
}

func (cb *CommitBlock) BlockTxs() []*txs.Tx { return nil }

func (cb *CommitBlock) Visit(v Visitor) error {
	return v.CommitBlock(cb)
}

func NewCommitBlock(
	parentID ids.ID,
	height uint64,
) (*CommitBlock, error) {
	res := &CommitBlock{
		CommonBlock: CommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
	}

	// We serialize this block as a Block so that it can be deserialized into a
	// Block
	blk := Block(res)
	bytes, err := Codec.Marshal(txs.Version, &blk)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
	}

	return res, res.initialize(bytes)
}
