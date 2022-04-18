// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package getter

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/metric"
)

// Get requests are always served, regardless node state (bootstrapping or normal operations).
var _ common.AllGetsServer = &getter{}

func New(vm block.ChainVM, commonCfg common.Config) (common.AllGetsServer, error) {
	gh := &getter{
		vm:     vm,
		sender: commonCfg.Sender,
		cfg:    commonCfg,
		log:    commonCfg.Ctx.Log,
	}

	ssVM, ok := gh.vm.(block.StateSyncableVM)
	if ok {
		gh.ssVM = ssVM
	}

	var err error
	gh.getAncestorsBlks, err = metric.NewAverager(
		"bs",
		"get_ancestors_blks",
		"blocks fetched in a call to GetAncestors",
		commonCfg.Ctx.Registerer,
	)
	return gh, err
}

type getter struct {
	vm     block.ChainVM
	ssVM   block.StateSyncableVM
	sender common.Sender
	cfg    common.Config

	log              logging.Logger
	getAncestorsBlks metric.Averager
}

func (gh *getter) GetStateSummaryFrontier(validatorID ids.ShortID, requestID uint32) error {
	if gh.ssVM == nil {
		gh.log.Debug("State sync not supported. GetStateSummaryFrontier(%s, %d) dropped.", validatorID, requestID)
		return nil
	}

	summary, err := gh.ssVM.StateSyncGetLastSummary()
	if err == common.ErrStateSyncableVMNotImplemented {
		// this may happen with rpcchainVMs
		gh.log.Debug("State sync not supported. GetStateSummaryFrontier(%s, %d) dropped.", validatorID, requestID)
		return nil
	} else if err != nil {
		gh.log.Info("couldn't get state summary frontier with %s. Dropping GetStateSummaryFrontier(%s, %d)",
			err, validatorID, requestID)
		return nil
	}
	gh.sender.SendStateSummaryFrontier(validatorID, requestID, summary.Bytes())
	return nil
}

func (gh *getter) GetAcceptedStateSummary(validatorID ids.ShortID, requestID uint32, keys []uint64) error {
	if gh.ssVM == nil {
		gh.log.Debug("State sync not supported. GetAcceptedStateSummary(%s, %d) dropped.", validatorID, requestID)
		return nil
	}

	summaryIDs := make([]ids.ID, 0, len(keys))
	for _, key := range keys {
		summary, err := gh.ssVM.StateSyncGetSummary(key)
		if err == common.ErrStateSyncableVMNotImplemented {
			// this may happen with rpcchainVMs
			return nil
		} else if err != nil {
			continue
		}
		summaryIDs = append(summaryIDs, summary.ID())
	}
	gh.sender.SendAcceptedStateSummary(validatorID, requestID, summaryIDs)
	return nil
}

func (gh *getter) GetAcceptedFrontier(validatorID ids.ShortID, requestID uint32) error {
	lastAccepted, err := gh.vm.LastAccepted()
	if err != nil {
		return err
	}
	gh.sender.SendAcceptedFrontier(validatorID, requestID, []ids.ID{lastAccepted})
	return nil
}

func (gh *getter) GetAccepted(validatorID ids.ShortID, requestID uint32, containerIDs []ids.ID) error {
	acceptedIDs := make([]ids.ID, 0, len(containerIDs))
	for _, blkID := range containerIDs {
		if blk, err := gh.vm.GetBlock(blkID); err == nil && blk.Status() == choices.Accepted {
			acceptedIDs = append(acceptedIDs, blkID)
		}
	}
	gh.sender.SendAccepted(validatorID, requestID, acceptedIDs)
	return nil
}

func (gh *getter) GetAncestors(validatorID ids.ShortID, requestID uint32, blkID ids.ID) error {
	ancestorsBytes, err := block.GetAncestors(
		gh.vm,
		blkID,
		gh.cfg.AncestorsMaxContainersSent,
		constants.MaxContainersLen,
		gh.cfg.MaxTimeGetAncestors,
	)
	if err != nil {
		gh.log.Verbo("couldn't get ancestors with %s. Dropping GetAncestors(%s, %d, %s)",
			err, validatorID, requestID, blkID)
		return nil
	}

	gh.getAncestorsBlks.Observe(float64(len(ancestorsBytes)))
	gh.sender.SendAncestors(validatorID, requestID, ancestorsBytes)
	return nil
}

func (gh *getter) Get(validatorID ids.ShortID, requestID uint32, blkID ids.ID) error {
	blk, err := gh.vm.GetBlock(blkID)
	if err != nil {
		// If we failed to get the block, that means either an unexpected error
		// has occurred, [vdr] is not following the protocol, or the
		// block has been pruned.
		gh.log.Debug("Get(%s, %d, %s) failed with: %s", validatorID, requestID, blkID, err)
		return nil
	}

	// Respond to the validator with the fetched block and the same requestID.
	gh.sender.SendPut(validatorID, requestID, blkID, blk.Bytes())
	return nil
}
