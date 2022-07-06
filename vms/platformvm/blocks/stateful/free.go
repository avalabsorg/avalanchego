// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

var _ freer = &freerImpl{}

type freer interface {
	freeProposalBlock(b *ProposalBlock)
	freeCommonBlock(b *commonBlock)
	freeAtomicBlock(b *AtomicBlock)
	freeCommitBlock(b *CommitBlock)
	freeAbortBlock(b *AbortBlock)
	freeStandardBlock(b *StandardBlock)
}

type freerImpl struct {
	backend
}

func (f *freerImpl) freeProposalBlock(b *ProposalBlock) {
	f.freeCommonBlock(b.commonBlock)
	b.onCommitState = nil
	b.onAbortState = nil
}

func (f *freerImpl) freeAtomicBlock(b *AtomicBlock) {
	f.freeDecisionBlock(b.decisionBlock)
}

func (f *freerImpl) freeAbortBlock(b *AbortBlock) {
	f.freeDecisionBlock(b.decisionBlock)
}

func (f *freerImpl) freeCommitBlock(b *CommitBlock) {
	f.freeDecisionBlock(b.decisionBlock)
}

func (f *freerImpl) freeStandardBlock(b *StandardBlock) {
	f.freeDecisionBlock(b.decisionBlock)
}

func (f *freerImpl) freeDecisionBlock(b *decisionBlock) {
	f.freeCommonBlock(b.commonBlock)
	b.onAcceptState = nil
	b.onAcceptFunc = nil
}

func (f *freerImpl) freeCommonBlock(b *commonBlock) {
	f.unpinVerifiedBlock(b.baseBlk.ID())
	b.children = nil
}
