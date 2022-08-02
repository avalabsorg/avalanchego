// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"go.uber.org/zap"
)

var _ stateless.Visitor = &rejector{}

// rejector handles the logic for rejecting a block.
type rejector struct {
	*backend
}

func (r *rejector) BlueberryProposalBlock(b *stateless.BlueberryProposalBlock) error {
	return r.visitProposalBlock(b)
}

func (r *rejector) ApricotProposalBlock(b *stateless.ApricotProposalBlock) error {
	return r.visitProposalBlock(b)
}

func (r *rejector) visitProposalBlock(b stateless.Block) error {
	blkID := b.ID()
	defer r.free(blkID)

	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", "proposal"),
		zap.Stringer("blkID", blkID),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parent", b.Parent()),
	)

	tx := b.BlockTxs()[0]
	if err := r.Mempool.Add(tx); err != nil {
		r.ctx.Log.Verbo(
			"failed to reissue tx",
			zap.Stringer("txID", tx.ID()),
			zap.Stringer("blkID", blkID),
			zap.Error(err),
		)
	}

	r.stateVersions.DeleteState(blkID)
	r.state.AddStatelessBlock(b, choices.Rejected)
	return r.state.Commit()
}

func (r *rejector) AtomicBlock(b *stateless.AtomicBlock) error {
	blkID := b.ID()
	defer r.free(blkID)

	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", "atomic"),
		zap.Stringer("blkID", blkID),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parent", b.Parent()),
	)

	if err := r.Mempool.Add(b.Tx); err != nil {
		r.ctx.Log.Debug(
			"failed to reissue tx",
			zap.Stringer("txID", b.Tx.ID()),
			zap.Stringer("blkID", blkID),
			zap.Error(err),
		)
	}

	r.stateVersions.DeleteState(blkID)
	r.state.AddStatelessBlock(b, choices.Rejected)
	return r.state.Commit()
}

func (r *rejector) BlueberryStandardBlock(b *stateless.BlueberryStandardBlock) error {
	return r.visitStandardBlock(b)
}

func (r *rejector) ApricotStandardBlock(b *stateless.ApricotStandardBlock) error {
	return r.visitStandardBlock(b)
}

func (r *rejector) visitStandardBlock(b stateless.Block) error {
	blkID := b.ID()
	defer r.free(blkID)

	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", "standard"),
		zap.Stringer("blkID", blkID),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parent", b.Parent()),
	)

	txes := b.BlockTxs()
	for _, tx := range txes {
		if err := r.Mempool.Add(tx); err != nil {
			r.ctx.Log.Debug(
				"failed to reissue tx",
				zap.Stringer("txID", tx.ID()),
				zap.Stringer("blkID", blkID),
				zap.Error(err),
			)
		}
	}

	r.stateVersions.DeleteState(blkID)
	r.state.AddStatelessBlock(b, choices.Rejected)
	return r.state.Commit()
}

func (r *rejector) CommitBlock(b *stateless.CommitBlock) error {
	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", "commit"),
		zap.Stringer("blkID", b.ID()),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parent", b.Parent()),
	)
	return r.rejectOptionBlock(b)
}

func (r *rejector) AbortBlock(b *stateless.AbortBlock) error {
	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", "abort"),
		zap.Stringer("blkID", b.ID()),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parent", b.Parent()),
	)
	return r.rejectOptionBlock(b)
}

func (r *rejector) rejectOptionBlock(b stateless.Block) error {
	blkID := b.ID()
	defer r.free(blkID)

	r.stateVersions.DeleteState(blkID)
	r.state.AddStatelessBlock(b, choices.Rejected)
	return r.state.Commit()
}
