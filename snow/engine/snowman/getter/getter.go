// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package getter

import (
	"context"

	"go.uber.org/zap"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/metric"
)

// Get requests are always served, regardless node state (bootstrapping or normal operations).
var _ common.AllGetsServer = &getter{}

func New(
	vm block.ChainVM,
	commonCfg common.Config,
) (common.AllGetsServer, error) {
	ssVM, _ := vm.(block.StateSyncableVM)
	gh := &getter{
		vm:     vm,
		ssVM:   ssVM,
		sender: commonCfg.Sender,
		cfg:    commonCfg,
		log:    commonCfg.Ctx.Log,
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
	ssVM   block.StateSyncableVM // can be nil
	sender common.Sender
	cfg    common.Config

	log              logging.Logger
	getAncestorsBlks metric.Averager
}

func (gh *getter) GetStateSummaryFrontier(parentCtx context.Context, nodeID ids.NodeID, requestID uint32) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.GetStateSummaryFrontier")
	defer span.End()

	// Note: we do not check if gh.ssVM.StateSyncEnabled since we want all
	// nodes, including those disabling state sync to serve state summaries if
	// these are available
	if gh.ssVM == nil {
		gh.log.Debug("dropping GetStateSummaryFrontier message",
			zap.String("reason", "state sync not supported"),
			zap.Stringer("nodeID", nodeID),
			zap.Uint32("requestID", requestID),
		)
		return nil
	}

	summary, err := gh.ssVM.GetLastStateSummary()
	if err != nil {
		gh.log.Debug("dropping GetStateSummaryFrontier message",
			zap.String("reason", "couldn't get state summary frontier"),
			zap.Stringer("nodeID", nodeID),
			zap.Uint32("requestID", requestID),
			zap.Error(err),
		)
		return nil
	}

	gh.sender.SendStateSummaryFrontier(ctx, nodeID, requestID, summary.Bytes())
	return nil
}

func (gh *getter) GetAcceptedStateSummary(parentCtx context.Context, nodeID ids.NodeID, requestID uint32, heights []uint64) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.GetAcceptedStateSummary")
	defer span.End()

	// If there are no requested heights, then we can return the result
	// immediately, regardless of if the underlying VM implements state sync.
	if len(heights) == 0 {
		gh.sender.SendAcceptedStateSummary(ctx, nodeID, requestID, nil)
		return nil
	}

	// Note: we do not check if gh.ssVM.StateSyncEnabled since we want all
	// nodes, including those disabling state sync to serve state summaries if
	// these are available
	if gh.ssVM == nil {
		gh.log.Debug("dropping GetAcceptedStateSummary message",
			zap.String("reason", "state sync not supported"),
			zap.Stringer("nodeID", nodeID),
			zap.Uint32("requestID", requestID),
		)
		return nil
	}

	summaryIDs := make([]ids.ID, 0, len(heights))
	for _, height := range heights {
		summary, err := gh.ssVM.GetStateSummary(height)
		if err == block.ErrStateSyncableVMNotImplemented {
			gh.log.Debug("dropping GetAcceptedStateSummary message",
				zap.String("reason", "state sync not supported"),
				zap.Stringer("nodeID", nodeID),
				zap.Uint32("requestID", requestID),
			)
			return nil
		}
		if err != nil {
			gh.log.Debug("couldn't get state summary",
				zap.Uint64("height", height),
				zap.Error(err),
			)
			continue
		}
		summaryIDs = append(summaryIDs, summary.ID())
	}

	gh.sender.SendAcceptedStateSummary(ctx, nodeID, requestID, summaryIDs)
	return nil
}

func (gh *getter) GetAcceptedFrontier(parentCtx context.Context, nodeID ids.NodeID, requestID uint32) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.GetAcceptedFrontier")
	defer span.End()

	_, lastAcceptedSpan := trace.Tracer().Start(ctx, "GetLastAccepted")
	lastAccepted, err := gh.vm.LastAccepted()
	lastAcceptedSpan.End()
	if err != nil {
		return err
	}
	gh.sender.SendAcceptedFrontier(ctx, nodeID, requestID, []ids.ID{lastAccepted})
	return nil
}

func (gh *getter) GetAccepted(parentCtx context.Context, nodeID ids.NodeID, requestID uint32, containerIDs []ids.ID) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.GetAccepted")
	defer span.End()

	acceptedIDs := make([]ids.ID, 0, len(containerIDs))
	for _, blkID := range containerIDs {
		_, getBlockSpan := trace.Tracer().Start(ctx, "GetBlock")
		blk, err := gh.vm.GetBlock(blkID)
		getBlockSpan.End()
		if err == nil && blk.Status() == choices.Accepted {
			acceptedIDs = append(acceptedIDs, blkID)
		}
	}
	gh.sender.SendAccepted(ctx, nodeID, requestID, acceptedIDs)
	return nil
}

func (gh *getter) GetAncestors(parentCtx context.Context, nodeID ids.NodeID, requestID uint32, blkID ids.ID) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.GetAncestors")
	defer span.End()

	ancestorsBytes, err := block.GetAncestors(
		gh.vm,
		blkID,
		gh.cfg.AncestorsMaxContainersSent,
		constants.MaxContainersLen,
		gh.cfg.MaxTimeGetAncestors,
	)
	if err != nil {
		gh.log.Verbo("dropping GetAncestors message",
			zap.String("reason", "couldn't get ancestors"),
			zap.Stringer("nodeID", nodeID),
			zap.Uint32("requestID", requestID),
			zap.Stringer("blkID", blkID),
			zap.Error(err),
		)
		return nil
	}

	gh.getAncestorsBlks.Observe(float64(len(ancestorsBytes)))
	gh.sender.SendAncestors(ctx, nodeID, requestID, ancestorsBytes)
	return nil
}

func (gh *getter) Get(parentCtx context.Context, nodeID ids.NodeID, requestID uint32, blkID ids.ID) error {
	ctx, span := trace.Tracer().Start(parentCtx, "getter.Get")
	defer span.End()

	_, getBlockSpan := trace.Tracer().Start(ctx, "GetBlock")
	blk, err := gh.vm.GetBlock(blkID)
	getBlockSpan.End()
	if err != nil {
		// If we failed to get the block, that means either an unexpected error
		// has occurred, [vdr] is not following the protocol, or the
		// block has been pruned.
		gh.log.Debug("failed Get request",
			zap.Stringer("nodeID", nodeID),
			zap.Uint32("requestID", requestID),
			zap.Stringer("blkID", blkID),
			zap.Error(err),
		)
		return nil
	}

	// Respond to the validator with the fetched block and the same requestID.
	gh.sender.SendPut(ctx, nodeID, requestID, blkID, blk.Bytes())
	return nil
}
