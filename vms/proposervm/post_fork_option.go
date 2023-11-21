// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package proposervm

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	smblock "github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
)

var _ PostForkBlock = (*postForkOption)(nil)

// The parent of a *postForkOption must be a *postForkBlock.
type postForkOption struct {
	block.Block
	postForkCommonComponents

	timestamp time.Time
}

func (b *postForkOption) Timestamp() time.Time {
	if b.Status() == choices.Accepted {
		return b.vm.lastAcceptedTime
	}
	return b.timestamp
}

func (b *postForkOption) Accept(ctx context.Context) error {
	if err := b.acceptOuterBlk(); err != nil {
		return err
	}
	return b.acceptInnerBlk(ctx)
}

func (b *postForkOption) acceptOuterBlk() error {
	// Update in-memory references
	b.status = choices.Accepted

	return b.vm.acceptPostForkBlock(b)
}

func (b *postForkOption) acceptInnerBlk(ctx context.Context) error {
	// mark the inner block as accepted and all conflicting inner blocks as
	// rejected
	return b.vm.Tree.Accept(ctx, b.innerBlk)
}

func (b *postForkOption) Reject(context.Context) error {
	// we do not reject the inner block here because that block may be contained
	// in the proposer block that causing this block to be rejected.
	blkID := b.ID()
	delete(b.vm.verifiedProposerBlocks, blkID)
	delete(b.vm.verifiedBlocks, blkID)
	b.status = choices.Rejected
	return nil
}

func (b *postForkOption) Status() choices.Status {
	if b.status == choices.Accepted && b.Height() > b.vm.lastAcceptedHeight {
		return choices.Processing
	}
	return b.status
}

func (b *postForkOption) Parent() ids.ID {
	return b.ParentID()
}

func (b *postForkOption) VerifyProposer(ctx context.Context) error {
	parent, err := b.vm.getBlock(ctx, b.ParentID())
	if err != nil {
		return err
	}
	return parent.verifyProposerPostForkOption(ctx, b)
}

// If Verify returns nil, Accept or Reject is eventually called on [b] and
// [b.innerBlk].
func (b *postForkOption) Verify(ctx context.Context) error {
	parent, err := b.vm.getBlock(ctx, b.ParentID())
	if err != nil {
		return err
	}
	b.timestamp = parent.Timestamp()
	return parent.verifyPostForkOption(ctx, b)
}

func (*postForkOption) verifyPreForkChild(context.Context, *preForkBlock) error {
	return errUnsignedChild
}

func (b *postForkOption) verifyProposerPostForkChild(ctx context.Context, child *postForkBlock) error {
	parentTimestamp := b.Timestamp()
	parentPChainHeight, err := b.pChainHeight(ctx)
	if err != nil {
		return err
	}
	err = b.postForkCommonComponents.Verify(
		ctx,
		parentTimestamp,
		parentPChainHeight,
		child,
	)
	if err != nil {
		return err
	}

	childID := child.ID()
	child.vm.verifiedProposerBlocks[childID] = child
	return nil
}

func (b *postForkOption) verifyPostForkChild(ctx context.Context, child *postForkBlock) error {
	parentPChainHeight, err := b.pChainHeight(ctx)
	if err != nil {
		return err
	}
	return child.vm.verifyAndRecordInnerBlk(
		ctx,
		&smblock.Context{
			PChainHeight: parentPChainHeight,
		},
		child,
	)
}

// A *postForkOption's parent can't be a *postForkOption
func (*postForkOption) verifyProposerPostForkOption(context.Context, *postForkOption) error {
	return errUnexpectedBlockType
}

func (*postForkOption) verifyPostForkOption(context.Context, *postForkOption) error {
	return errUnexpectedBlockType
}

func (b *postForkOption) buildChild(ctx context.Context) (Block, error) {
	parentID := b.ID()
	parentPChainHeight, err := b.pChainHeight(ctx)
	if err != nil {
		b.vm.ctx.Log.Error("unexpected build block failure",
			zap.String("reason", "failed to fetch parent's P-chain height"),
			zap.Stringer("parentID", parentID),
			zap.Error(err),
		)
		return nil, err
	}
	return b.postForkCommonComponents.buildChild(
		ctx,
		parentID,
		b.Timestamp(),
		parentPChainHeight,
	)
}

// This block's P-Chain height is its parent's P-Chain height
func (b *postForkOption) pChainHeight(ctx context.Context) (uint64, error) {
	parent, err := b.vm.getBlock(ctx, b.ParentID())
	if err != nil {
		return 0, err
	}
	return parent.pChainHeight(ctx)
}

func (b *postForkOption) setStatus(status choices.Status) {
	b.status = status
}

func (b *postForkOption) getStatelessBlk() block.Block {
	return b.Block
}
