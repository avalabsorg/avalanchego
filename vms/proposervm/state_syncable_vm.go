// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package proposervm

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/vms/proposervm/summary"
)

var errStateSyncResults = errors.New("could not retrieve state sync results")

func (vm *VM) StateSyncEnabled() (bool, error) {
	if vm.innerStateSyncVM == nil {
		return false, nil
	}

	// if vm implements Snowman++, a block height index must be available
	// to support state sync
	if vm.VerifyHeightIndex() != nil {
		return false, nil
	}

	return vm.innerStateSyncVM.StateSyncEnabled()
}

func (vm *VM) GetOngoingSyncStateSummary() (common.Summary, error) {
	if vm.innerStateSyncVM == nil {
		return nil, common.ErrStateSyncableVMNotImplemented
	}

	innerSummary, err := vm.innerStateSyncVM.GetOngoingSyncStateSummary()
	if err != nil {
		return nil, err
	}

	var proSummary summary.ProposerSummaryIntf
	if innerSummary.ID() == ids.Empty {
		// summary with emptyID signals no local summary is available in InnerVM
		proSummary, err = summary.BuildEmptyProposerSummary(innerSummary)
	} else {
		proBlkID, err := vm.GetBlockIDAtHeight(innerSummary.Height())
		if err != nil {
			// this should never happen, it's proVM being out of sync with innerVM
			vm.ctx.Log.Warn("inner summary unknown to proposer VM. Block height index missing: %s", err)
			return nil, common.ErrUnknownStateSummary
		}
		proSummary, err = summary.BuildProposerSummary(proBlkID, innerSummary)
		if err != nil {
			return nil, err
		}
	}

	return &statefulSummary{
		ProposerSummaryIntf: proSummary,
		innerSummary:        innerSummary,
		vm:                  vm,
	}, err
}

func (vm *VM) GetLastStateSummary() (common.Summary, error) {
	if vm.innerStateSyncVM == nil {
		return nil, common.ErrStateSyncableVMNotImplemented
	}

	// Extract inner vm's last state summary
	innerSummary, err := vm.innerStateSyncVM.GetLastStateSummary()
	if err != nil {
		return nil, err // including common.ErrUnknownStateSummary case
	}

	// retrieve ProBlkID
	proBlkID, err := vm.GetBlockIDAtHeight(innerSummary.Height())
	if err != nil {
		// this should never happen, since that would mean proVM has become out of sync with innerVM
		vm.ctx.Log.Warn("inner summary unknown to proposer VM. Block height index missing: %s", err)
		return nil, common.ErrUnknownStateSummary
	}

	proSummary, err := summary.BuildProposerSummary(proBlkID, innerSummary)
	return &statefulSummary{
		ProposerSummaryIntf: proSummary,
		innerSummary:        innerSummary,
		vm:                  vm,
	}, err
}

// Note: it's important that ParseStateSummary do not use any index or state
// to allow summaries being parsed also by freshly started node with no previous state.
func (vm *VM) ParseStateSummary(summaryBytes []byte) (common.Summary, error) {
	if vm.innerStateSyncVM == nil {
		return nil, common.ErrStateSyncableVMNotImplemented
	}

	statelessSummary, err := summary.Parse(summaryBytes)
	if err != nil {
		return nil, err
	}

	innerSummary, err := vm.innerStateSyncVM.ParseStateSummary(statelessSummary.InnerBytes())
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal innerSummaryContent due to: %w", err)
	}

	return &statefulSummary{
		ProposerSummaryIntf: &summary.ProposerSummary{
			StatelessSummary: *statelessSummary.(*summary.StatelessSummary),
			SummaryHeight:    innerSummary.Height(),
		},
		innerSummary: innerSummary,
		vm:           vm,
	}, nil
}

func (vm *VM) GetStateSummary(height uint64) (common.Summary, error) {
	if vm.innerStateSyncVM == nil {
		return nil, common.ErrStateSyncableVMNotImplemented
	}

	innerSummary, err := vm.innerStateSyncVM.GetStateSummary(height)
	if err != nil {
		return nil, err // including common.ErrUnknownStateSummary case
	}

	// retrieve ProBlkID
	proBlkID, err := vm.GetBlockIDAtHeight(innerSummary.Height())
	if err != nil {
		// this should never happen, since that would mean proVM has become out of sync with innerVM
		vm.ctx.Log.Warn("inner summary unknown to proposer VM. Block height index missing: %s", err)
		return nil, common.ErrUnknownStateSummary
	}

	proSummary, err := summary.BuildProposerSummary(proBlkID, innerSummary)
	return &statefulSummary{
		ProposerSummaryIntf: proSummary,
		innerSummary:        innerSummary,
		vm:                  vm,
	}, err
}

func (vm *VM) GetStateSyncResult() (ids.ID, uint64, error) {
	if vm.innerStateSyncVM == nil {
		return ids.Empty, 0, common.ErrStateSyncableVMNotImplemented
	}

	_, height, err := vm.innerStateSyncVM.GetStateSyncResult()
	if err != nil {
		return ids.Empty, 0, err
	}
	proBlkID, err := vm.GetBlockIDAtHeight(height)
	if err != nil {
		return ids.Empty, 0, errStateSyncResults
	}
	return proBlkID, height, nil
}

func (vm *VM) ParseStateSyncableBlock(blkBytes []byte) (snowman.StateSyncableBlock, error) {
	if vm.innerStateSyncVM == nil {
		return nil, common.ErrStateSyncableVMNotImplemented
	}

	proposerBlk, err := vm.parseBlock(blkBytes)
	if err != nil {
		return nil, err
	}

	innerStateSyncableBlk, err := vm.innerStateSyncVM.ParseStateSyncableBlock(proposerBlk.getInnerBlk().Bytes())
	if err != nil {
		return nil, err
	}

	return &stateSyncableBlock{
		Block:                 proposerBlk,
		innerStateSyncableBlk: innerStateSyncableBlk,
	}, nil
}
