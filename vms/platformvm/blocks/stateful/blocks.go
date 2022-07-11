// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/chains/atomic"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
)

var (
	_ snowman.Block       = &block{}
	_ snowman.OracleBlock = &oracleBlock{}
)

func newBlock(
	blk stateless.Block,
	manager *manager,
) snowman.Block {
	b := &block{
		manager: manager,
		Block:   blk,
	}
	// TODO should we just have a NewOracleBlock method?
	if _, ok := blk.(*stateless.ProposalBlock); ok {
		return &oracleBlock{
			block: b,
		}
	}
	return b
}

type block struct {
	stateless.Block
	manager *manager
}

func (b *block) Verify() error {
	return b.Visit(b.manager.verifier)
}

func (b *block) Accept() error {
	return b.Visit(b.manager.acceptor)
}

func (b *block) Reject() error {
	return b.Visit(b.manager.rejector)
}

// TODO
func (b *block) Status() choices.Status {
	blkID := b.ID()
	if b.manager.state.GetLastAccepted() == blkID {
		return choices.Accepted
	}
	// Check if the block is in memory
	if _, ok := b.manager.backend.blkIDToState[blkID]; ok {
		return choices.Processing
	}
	// Block isn't in memory. Check in the database.
	_, status, err := b.manager.GetStatelessBlock(blkID)
	if err != nil {
		// It isn't in the database.
		// TODO is this right?
		return choices.Processing
	}
	return status
}

// TODO
func (b *block) Timestamp() time.Time {
	// 	 If this is the last accepted block and the block was loaded from disk
	// 	 since it was accepted, then the timestamp wouldn't be set correctly. So,
	// 	 we explicitly return the chain time.
	blkID := b.ID()
	// Check if the block is processing.
	if blkState, ok := b.manager.blkIDToState[blkID]; ok {
		return blkState.timestamp
	}
	// The block isn't processing.
	// According to the snowman.Block interface, the last accepted
	// block is the only accepted block that must return a correct timestamp,
	// so we just return the chain time.
	return b.manager.state.GetTimestamp()
}

type oracleBlock struct {
	// Invariant: The inner statless block is a *stateless.ProposalBlock.
	*block
}

func (b *oracleBlock) Options() ([2]snowman.Block, error) {
	blkID := b.ID()
	nextHeight := b.Height() + 1

	statelessCommitBlk, err := stateless.NewCommitBlock(
		blkID,
		nextHeight,
	)
	if err != nil {
		return [2]snowman.Block{}, fmt.Errorf(
			"failed to create commit block: %w",
			err,
		)
	}
	commitBlock := b.manager.NewBlock(statelessCommitBlk)

	statelessAbortBlk, err := stateless.NewAbortBlock(
		blkID,
		nextHeight,
	)
	if err != nil {
		return [2]snowman.Block{}, fmt.Errorf(
			"failed to create abort block: %w",
			err,
		)
	}
	abortBlock := b.manager.NewBlock(statelessAbortBlk)

	if b.manager.backend.blkIDToState[blkID].inititallyPreferCommit {
		return [2]snowman.Block{commitBlock, abortBlock}, nil
	}
	return [2]snowman.Block{abortBlock, commitBlock}, nil
}

type blockState struct {
	// TODO add stateless block to this struct
	statelessBlock         stateless.Block
	onAcceptFunc           func()
	onAcceptState          state.Diff
	onCommitState          state.Diff
	onAbortState           state.Diff
	children               []ids.ID
	timestamp              time.Time
	inputs                 ids.Set
	atomicRequests         map[ids.ID]*atomic.Requests
	inititallyPreferCommit bool
}
