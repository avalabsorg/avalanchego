// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/config"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/executor"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/mempool"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestVerifierVisitProposalBlock(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := state.NewMockState(ctrl)
	mempool := mempool.NewMockMempool(ctrl)
	parentID := ids.GenerateTestID()
	parentStatelessBlk := stateless.NewMockBlock(ctrl)
	verifier := &verifier{
		txExecutorBackend: executor.Backend{},
		backend: &backend{
			blkIDToState: map[ids.ID]*blockState{
				parentID: {
					statelessBlock: parentStatelessBlk,
				},
			},
			Mempool:       mempool,
			state:         s,
			stateVersions: state.NewVersions(parentID, s),
			ctx: &snow.Context{
				Log: logging.NoLog{},
			},
		},
	}

	onCommitState := state.NewMockDiff(ctrl)
	onAbortState := state.NewMockDiff(ctrl)
	blkTx := txs.NewMockUnsignedTx(ctrl)
	blkTx.EXPECT().Visit(gomock.AssignableToTypeOf(&executor.ProposalTxExecutor{})).DoAndReturn(
		func(e *executor.ProposalTxExecutor) error {
			e.OnCommit = onCommitState
			e.OnAbort = onAbortState
			return nil
		},
	).Times(1)
	blkTx.EXPECT().Initialize(gomock.Any()).Times(2)

	blk, err := stateless.NewProposalBlock(
		parentID,
		2,
		&txs.Tx{
			Unsigned: blkTx,
			Creds:    []verify.Verifiable{},
		},
	)
	assert.NoError(err)

	// Set expectations for dependencies.
	timestamp := time.Now()
	parentStatelessBlk.EXPECT().Height().Return(uint64(1)).Times(1)
	mempool.EXPECT().RemoveProposalTx(blk.Tx).Times(1)
	onCommitState.EXPECT().AddTx(blk.Tx, status.Committed).Times(1)
	onAbortState.EXPECT().AddTx(blk.Tx, status.Aborted).Times(1)
	onAbortState.EXPECT().GetTimestamp().Return(timestamp).Times(1)

	// Visit the block
	err = verifier.VisitProposalBlock(blk)
	assert.NoError(err)
	assert.Contains(verifier.backend.blkIDToState, blk.ID())
	gotBlkState := verifier.backend.blkIDToState[blk.ID()]
	assert.Equal(blk, gotBlkState.statelessBlock)
	assert.Equal(onCommitState, gotBlkState.onCommitState)
	assert.Equal(onAbortState, gotBlkState.onAbortState)
	assert.Equal(timestamp, gotBlkState.timestamp)
}

func TestVerifierVisitAtomicBlock(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := state.NewMockState(ctrl)
	mempool := mempool.NewMockMempool(ctrl)
	parentID := ids.GenerateTestID()
	parentStatelessBlk := stateless.NewMockBlock(ctrl)
	grandparentID := ids.GenerateTestID()
	stateVersions := state.NewMockVersions(ctrl)
	parentState := state.NewMockState(ctrl)
	verifier := &verifier{
		txExecutorBackend: executor.Backend{
			Config: &config.Config{
				ApricotPhase5Time: time.Now().Add(time.Hour),
			},
		},
		backend: &backend{
			blkIDToState: map[ids.ID]*blockState{
				parentID: {
					statelessBlock: parentStatelessBlk,
				},
			},
			Mempool:       mempool,
			state:         s,
			stateVersions: stateVersions,
			ctx: &snow.Context{
				Log: logging.NoLog{},
			},
		},
	}

	onAccept := state.NewMockDiff(ctrl)
	blkTx := txs.NewMockUnsignedTx(ctrl)
	blkTx.EXPECT().Visit(gomock.AssignableToTypeOf(&executor.AtomicTxExecutor{})).DoAndReturn(
		func(e *executor.AtomicTxExecutor) error {
			e.OnAccept = onAccept
			e.Inputs = ids.Set{}
			return nil
		},
	).Times(1)
	blkTx.EXPECT().Initialize(gomock.Any()).Times(2)

	blk, err := stateless.NewAtomicBlock(
		parentID,
		2,
		&txs.Tx{
			Unsigned: blkTx,
			Creds:    []verify.Verifiable{},
		},
	)
	assert.NoError(err)

	// Set expectations for dependencies.
	timestamp := time.Now()
	parentState.EXPECT().GetTimestamp().Return(timestamp).Times(1)
	stateVersions.EXPECT().GetState(blk.Parent()).Return(parentState, true).Times(1)
	parentStatelessBlk.EXPECT().Height().Return(uint64(1)).Times(1)
	parentStatelessBlk.EXPECT().Parent().Return(grandparentID).Times(1)
	s.EXPECT().GetStatelessBlock(parentID).Return(parentStatelessBlk, choices.Accepted, nil).Times(1)
	mempool.EXPECT().RemoveDecisionTxs([]*txs.Tx{blk.Tx}).Times(1)
	onAccept.EXPECT().AddTx(blk.Tx, status.Committed).Times(1)
	onAccept.EXPECT().GetTimestamp().Return(timestamp).Times(1)
	stateVersions.EXPECT().SetState(blk.ID(), onAccept).Times(1)

	err = verifier.VisitAtomicBlock(blk)
	assert.NoError(err)

	assert.Contains(verifier.backend.blkIDToState, blk.ID())
	gotBlkState := verifier.backend.blkIDToState[blk.ID()]
	assert.Equal(blk, gotBlkState.statelessBlock)
	assert.Equal(onAccept, gotBlkState.onAcceptState)
	assert.Equal(ids.Set{}, gotBlkState.inputs)
	assert.Equal(timestamp, gotBlkState.timestamp)
}
