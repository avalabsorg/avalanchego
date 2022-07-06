// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/mempool"
)

type heightSetter interface {
	SetHeight(height uint64)
}

// Shared fields used by visitors.
type backend struct {
	mempool.Mempool
	// TODO consolidate state fields below?
	blockState
	heightSetter
	// Block ID --> Function to be executed if the block is accepted.
	blkIDToOnAcceptFunc map[ids.ID]func()
	state               state.State
	ctx                 *snow.Context
	bootstrapped        *utils.AtomicBool
}

func (b *backend) getState() state.State {
	return b.state
}

// TODO do we even need this or can we just pass parent ID into getStatefulBlock?
func (b *backend) parent(blk *stateless.CommonBlock) (Block, error) {
	parentBlkID := blk.Parent()
	return b.GetStatefulBlock(parentBlkID)
}
