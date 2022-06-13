// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
)

// commonBlock contains fields and methods common to all full blocks in this VM.
type commonBlock struct {
	baseBlk   *stateless.CommonBlock
	timestamp time.Time // Time this block was proposed at
	status    choices.Status
	children  []Block

	verifier Verifier
}

func (c *commonBlock) parentBlock() (Block, error) {
	parentBlkID := c.baseBlk.Parent()
	return c.verifier.GetStatefulBlock(parentBlkID)
}

func (c *commonBlock) addChild(child Block) {
	c.children = append(c.children, child)
}

// Parent returns this block's parent's ID
func (c *commonBlock) Status() choices.Status { return c.status }

func (c *commonBlock) Timestamp() time.Time {
	// If this is the last accepted block and the block was loaded from disk
	// since it was accepted, then the timestamp wouldn't be set correctly. So,
	// we explicitly return the chain time.
	if c.baseBlk.ID() == c.verifier.GetLastAccepted() {
		return c.verifier.GetTimestamp()
	}
	return c.timestamp
}

func (c *commonBlock) conflicts(s ids.Set) (bool, error) {
	if c.Status() == choices.Accepted {
		return false, nil
	}
	parent, err := c.parentBlock()
	if err != nil {
		return false, err
	}
	return parent.conflicts(s)
}

func (c *commonBlock) verify() error {
	if c == nil {
		return ErrBlockNil
	}

	parent, err := c.parentBlock()
	if err != nil {
		return err
	}
	if expectedHeight := parent.Height() + 1; expectedHeight != c.baseBlk.Height() {
		return fmt.Errorf(
			"expected block to have height %d, but found %d",
			expectedHeight,
			c.baseBlk.Height(),
		)
	}
	return nil
}

func (c *commonBlock) free() {
	c.verifier.DropVerifiedBlock(c.baseBlk.ID())
	c.children = nil
}

func (c *commonBlock) accept() {
	blkID := c.baseBlk.ID()

	c.status = choices.Accepted
	c.verifier.SetLastAccepted(blkID)
	c.verifier.SetHeight(c.baseBlk.Height())
	c.verifier.AddToRecentlyAcceptedWindows(blkID)
}

func (c *commonBlock) reject() {
	defer c.free()
	c.status = choices.Rejected
}
