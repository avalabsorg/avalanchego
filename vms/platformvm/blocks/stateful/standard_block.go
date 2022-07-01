// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"errors"

	"github.com/ava-labs/avalanchego/chains/atomic"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/executor"
)

var (
	errConflictingBatchTxs = errors.New("block contains conflicting transactions")

	_ Block    = &StandardBlock{}
	_ Decision = &StandardBlock{}
)

// StandardBlock being accepted results in the transactions contained in the
// block to be accepted and committed to the chain.
type StandardBlock struct {
	Manager
	*stateless.StandardBlock
	*decisionBlock

	// Inputs are the atomic Inputs that are consumed by this block's atomic
	// transactions
	Inputs ids.Set

	atomicRequests map[ids.ID]*atomic.Requests
}

// NewStandardBlock returns a new *StandardBlock where the block's parent, a
// decision block, has ID [parentID].
func NewStandardBlock(
	manager Manager,
	txExecutorBackend executor.Backend,
	parentID ids.ID,
	height uint64,
	txs []*txs.Tx,
) (*StandardBlock, error) {
	statelessBlk, err := stateless.NewStandardBlock(parentID, height, txs)
	if err != nil {
		return nil, err
	}
	return toStatefulStandardBlock(statelessBlk, manager, txExecutorBackend, choices.Processing)
}

func toStatefulStandardBlock(
	statelessBlk *stateless.StandardBlock,
	manager Manager,
	txExecutorBackend executor.Backend,
	status choices.Status,
) (*StandardBlock, error) {
	sb := &StandardBlock{
		StandardBlock: statelessBlk,
		Manager:       manager,
		decisionBlock: &decisionBlock{
			chainState: manager,
			commonBlock: &commonBlock{
				timestampGetter:   manager,
				lastAccepteder:    manager,
				baseBlk:           &statelessBlk.CommonBlock,
				status:            status,
				txExecutorBackend: txExecutorBackend,
			},
		},
	}

	for _, tx := range sb.Txs {
		tx.Unsigned.InitCtx(sb.txExecutorBackend.Ctx)
	}

	return sb, nil
}

// conflicts checks to see if the provided input set contains any conflicts with
// any of this block's non-accepted ancestors or itself.
func (sb *StandardBlock) conflicts(s ids.Set) (bool, error) {
	return sb.conflictsStandardBlock(sb, s)
	/* TODO remove
	if sb.status == choices.Accepted {
		return false, nil
	}
	if sb.Inputs.Overlaps(s) {
		return true, nil
	}
	parent, err := sb.parentBlock()
	if err != nil {
		return false, err
	}
	return parent.conflicts(s)
	*/
}

// Verify this block performs a valid state transition.
//
// The parent block must be a proposal
//
// This function also sets onAcceptDB database if the verification passes.
func (sb *StandardBlock) Verify() error {
	return sb.verifyStandardBlock(sb)
}

func (sb *StandardBlock) Accept() error {
	return sb.acceptStandardBlock(sb)
}

func (sb *StandardBlock) Reject() error {
	return sb.rejectStandardBlock(sb)
}

func (sb *StandardBlock) free() {
	sb.freeStandardBlock(sb)
}

func (a *StandardBlock) setBaseState() {
	a.Manager.setBaseStateStandardBlock(a)
}
