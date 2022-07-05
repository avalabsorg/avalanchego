// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import "github.com/ava-labs/avalanchego/snow/choices"

var _ rejector = &rejectorImpl{}

type rejector interface {
	rejectProposalBlock(b *ProposalBlock) error
	rejectAtomicBlock(b *AtomicBlock) error
	rejectStandardBlock(b *StandardBlock) error
	rejectCommitBlock(b *CommitBlock) error
	rejectAbortBlock(b *AbortBlock) error
}

type rejectorImpl struct {
	backend
	freer
}

func (r *rejectorImpl) rejectProposalBlock(b *ProposalBlock) error {
	r.ctx.Log.Verbo(
		"Rejecting Proposal Block %s at height %d with parent %s",
		b.ID(),
		b.Height(),
		b.Parent(),
	)

	b.onCommitState = nil
	b.onAbortState = nil

	if err := r.Add(b.Tx); err != nil {
		r.ctx.Log.Verbo(
			"failed to reissue tx %q due to: %s",
			b.Tx.ID(),
			err,
		)
	}

	defer r.commonReject(b.commonBlock)
	r.AddStatelessBlock(b.ProposalBlock, b.Status())
	return r.Commit()
}

func (r *rejectorImpl) rejectAtomicBlock(b *AtomicBlock) error {
	r.ctx.Log.Verbo(
		"Rejecting Atomic Block %s at height %d with parent %s",
		b.ID(),
		b.Height(),
		b.Parent(),
	)

	if err := r.Add(b.Tx); err != nil {
		r.ctx.Log.Debug(
			"failed to reissue tx %q due to: %s",
			b.Tx.ID(),
			err,
		)
	}

	defer r.commonReject(b.commonBlock)
	r.AddStatelessBlock(b.AtomicBlock, b.Status())
	return r.Commit()
}

func (r *rejectorImpl) rejectStandardBlock(b *StandardBlock) error {
	r.ctx.Log.Verbo(
		"Rejecting Standard Block %s at height %d with parent %s",
		b.ID(),
		b.Height(),
		b.Parent(),
	)

	for _, tx := range b.Txs {
		if err := r.Add(tx); err != nil {
			r.ctx.Log.Debug(
				"failed to reissue tx %q due to: %s",
				tx.ID(),
				err,
			)
		}
	}

	defer r.commonReject(b.commonBlock)
	r.AddStatelessBlock(b.StandardBlock, b.Status())
	return r.Commit()
}

func (r *rejectorImpl) rejectCommitBlock(b *CommitBlock) error {
	r.ctx.Log.Verbo(
		"Rejecting CommitBlock Block %s at height %d with parent %s",
		b.ID(), b.Height(), b.Parent(),
	)

	defer r.commonReject(b.commonBlock)
	r.AddStatelessBlock(b.CommitBlock, b.Status())
	return r.Commit()
}

func (r *rejectorImpl) rejectAbortBlock(b *AbortBlock) error {
	r.ctx.Log.Verbo(
		"Rejecting Abort Block %s at height %d with parent %s",
		b.ID(),
		b.Height(),
		b.Parent(),
	)

	defer r.commonReject(b.commonBlock)
	r.AddStatelessBlock(b.AbortBlock, b.Status())
	return r.Commit()
}

func (r *rejectorImpl) commonReject(b *commonBlock) {
	b.status = choices.Rejected
	// We do the following rather than implementing free()
	// on *commonBlock so that in the future, we don't accidentally
	// forget to implement free() on a block type that embeds *commonBlock.
	r.freeCommonBlock(b)
}
