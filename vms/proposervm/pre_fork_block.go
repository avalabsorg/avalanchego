// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package proposervm

import (
	"errors"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
)

var (
	errExpectedNoProposer = errors.New("expected no proposer to be named")

	_ Block = &preForkBlock{}
)

// preForkBlock implements proposervm.Block
type preForkBlock struct {
	snowman.Block
	vm *VM
}

func (b *preForkBlock) Parent() ids.ID {
	return b.Block.Parent()
}

func (b *preForkBlock) Verify() error {
	b.vm.ctx.Log.Debug("Snowman++ calling verify on %s", b.ID())

	parent, err := b.vm.getBlock(b.Block.Parent())
	if err != nil {
		return err
	}
	return parent.verifyPreForkChild(b)
}

func (b *preForkBlock) Options() ([2]snowman.Block, error) {
	oracleBlk, ok := b.Block.(snowman.OracleBlock)
	if !ok {
		return [2]snowman.Block{}, snowman.ErrNotOracle
	}

	options, err := oracleBlk.Options()
	if err != nil {
		return [2]snowman.Block{}, err
	}
	// A pre-fork block's child options are always pre-fork blocks
	return [2]snowman.Block{
		&preForkBlock{
			Block: options[0],
			vm:    b.vm,
		},
		&preForkBlock{
			Block: options[1],
			vm:    b.vm,
		},
	}, nil
}

func (b *preForkBlock) verifyPreForkChild(child *preForkBlock) error {
	parentTimestamp := b.Timestamp()
	if !parentTimestamp.Before(b.vm.activationTime) {
		// If this block's timestamp is at or after activation time,
		// it's child must be a post-fork block
		return errProposersActivated
	}

	return child.Block.Verify()
}

func (b *preForkBlock) getInnerBlk() snowman.Block {
	return b.Block
}

// This method only returns nil once (during the transition)
func (b *preForkBlock) verifyPostForkChild(child *postForkBlock) error {
	childID := child.ID()
	childPChainHeight := child.PChainHeight()
	currentPChainHeight, err := b.vm.PChainHeight()
	if err != nil {
		b.vm.ctx.Log.Error("couldn't retrieve current P-Chain height while verifying %s: %s", childID, err)
		return err
	}
	if childPChainHeight > currentPChainHeight {
		b.vm.ctx.Log.Warn("Snowman++ verify - dropped post-fork block; expected chid's P-Chain height to be <=%d but got %d",
			currentPChainHeight, childPChainHeight)
		return errPChainHeightNotReached
	}

	// Make sure [b] is the parent of [child]'s inner block
	expectedInnerParentID := b.ID()
	innerParentID := child.innerBlk.Parent()
	if innerParentID != expectedInnerParentID {
		b.vm.ctx.Log.Warn("Snowman++ verify - dropped post-fork block; expected inner parent %s but got %s",
			expectedInnerParentID, innerParentID)
		return errInnerParentMismatch
	}

	// A *preForkBlock can only have a *postForkBlock child
	// if the *preForkBlock is the last *preForkBlock before activation takes effect
	// (its timestamp is at or after the activation time)
	parentTimestamp := b.Timestamp()
	if parentTimestamp.Before(b.vm.activationTime) {
		return errProposersNotActivated
	}

	// Child's timestamp must be at or after its parent's timestamp
	childTimestamp := child.Timestamp()
	if childTimestamp.Before(parentTimestamp) {
		b.vm.ctx.Log.Warn("Snowman++ verify - dropped post-fork block; expected child's timestamp (%s) to be at or after parent's timestamp (%s)",
			childTimestamp, parentTimestamp)
		return errTimeNotMonotonic
	}

	// Child timestamp can't be too far in the future
	maxTimestamp := b.vm.Time().Add(maxSkew)
	if childTimestamp.After(maxTimestamp) {
		b.vm.ctx.Log.Warn("Snowman++ verify - dropped post-fork block; block's timestamp (%s) is after the synchrony bound (%s)",
			childTimestamp, maxTimestamp)
		return errTimeTooAdvanced
	}

	proposer := child.Block.Proposer()
	if proposer != ids.ShortEmpty {
		return errExpectedNoProposer
	}

	// Verify the lack of signature on the node
	if err := child.Block.Verify(); err != nil {
		return err
	}

	// If inner block's Verify returned true, don't call it again.
	// Note that if [child.innerBlk.Verify] returns nil,
	// this method returns nil. This must always remain the case to
	// maintain the inner block's invariant that if it's Verify()
	// returns nil, it is eventually accepted/rejected.
	if !b.vm.Tree.Contains(child.innerBlk) {
		if err := child.innerBlk.Verify(); err != nil {
			return err
		}
		b.vm.Tree.Add(child.innerBlk)
	}

	b.vm.verifiedBlocks[childID] = child
	return nil
}

func (b *preForkBlock) verifyPostForkOption(child *postForkOption) error {
	b.vm.ctx.Log.Debug("post-fork option has pre-fork block as parent")
	return errUnexpectedBlockType
}

func (b *preForkBlock) buildChild(innerBlock snowman.Block) (Block, error) {
	parentTimestamp := b.Timestamp()
	if parentTimestamp.Before(b.vm.activationTime) {
		// The chain hasn't forked yet
		res := &preForkBlock{
			Block: innerBlock,
			vm:    b.vm,
		}
		b.vm.ctx.Log.Debug("Snowman++ build pre-fork block %s - timestamp parent block %v",
			res.ID(), b.Timestamp())

		return res, nil
	}

	// The chain is currently forking

	parentID := b.ID()
	newTimestamp := b.vm.Time().Truncate(time.Second)
	if newTimestamp.Before(parentTimestamp) {
		newTimestamp = parentTimestamp
	}

	pChainHeight, err := b.vm.PChainHeight()
	if err != nil {
		return nil, err
	}

	statelessBlock, err := block.BuildUnsigned(
		parentID,
		newTimestamp,
		pChainHeight,
		innerBlock.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	blk := &postForkBlock{
		Block: statelessBlock,
		postForkCommonComponents: postForkCommonComponents{
			vm:       b.vm,
			innerBlk: innerBlock,
			status:   choices.Processing,
		},
	}

	b.vm.ctx.Log.Debug("Snowman++ build post-fork block %s - parent timestamp %v, expected delay NA, block timestamp %v.",
		blk.ID(), parentTimestamp, newTimestamp)
	return blk, b.vm.storePostForkBlock(blk)
}

func (b *preForkBlock) pChainHeight() (uint64, error) {
	return 0, nil
}
