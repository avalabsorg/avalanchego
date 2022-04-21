// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package syncer

import (
	"fmt"
	"math"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/common/tracker"
	"github.com/stretchr/testify/assert"
)

func TestAtStateSyncDoneLastSummaryBlockIsRequested(t *testing.T) {
	assert := assert.New(t)

	vdrs := buildTestPeers(t)
	startupAlpha := (3*vdrs.Weight() + 3) / 4
	commonCfg := common.Config{
		Ctx:                         snow.DefaultConsensusContextTest(),
		Beacons:                     vdrs,
		SampleK:                     vdrs.Len(),
		Alpha:                       (vdrs.Weight() + 1) / 2,
		WeightTracker:               tracker.NewWeightTracker(vdrs, startupAlpha),
		RetryBootstrap:              true, // this sets RetryStateSyncing too
		RetryBootstrapWarnFrequency: 1,    // this sets RetrySyncingWarnFrequency too
	}
	syncer, fullVM, sender := buildTestsObjects(t, &commonCfg)

	stateSyncFullyDone := false
	syncer.onDoneStateSyncing = func(lastReqID uint32) error {
		stateSyncFullyDone = true
		return nil
	}

	// mock VM to return lastSummaryBlkID and be able to receive full block
	syncer.lastSummaryBlkID = ids.ID{'b', 'l', 'k', 'I', 'D'}
	fullVM.CantGetStateSyncResult = true
	fullVM.GetStateSyncResultF = func() error { return nil }

	successfulParseSyncableBlockBlkMock := func(
		b []byte,
	) (snowman.StateSyncableBlock, error) {
		return &snowman.TestStateSyncableBlock{
			TestBlock: snowman.TestBlock{
				TestDecidable: choices.TestDecidable{
					IDV:     syncer.lastSummaryBlkID,
					StatusV: choices.Processing,
				},
				BytesV: b,
			},
			T:         t,
			RegisterF: func() error { return nil },
		}, nil
	}

	fullVM.CantParseStateSyncableBlock = true
	fullVM.ParseStateSyncableBlockF = successfulParseSyncableBlockBlkMock

	// mock sender to record requested blkID
	var (
		blkRequested  bool
		reqBlkID      ids.ID
		reachedNodeID = ids.ShortID{'n', 'o', 'd', 'e', 'I', 'D'}
		sentReqID     uint32
	)
	sender.CantSendGet = true
	sender.SendGetF = func(nodeID ids.ShortID, reqID uint32, blkID ids.ID) {
		blkRequested = true
		reachedNodeID = nodeID
		sentReqID = reqID
		reqBlkID = blkID
	}

	// Any Put response before StateSyncDone is received from VM is dropped
	assert.NoError(syncer.Put(reachedNodeID, sentReqID, []byte{}))
	assert.False(stateSyncFullyDone)

	assert.NoError(syncer.Notify(common.StateSyncDone))
	assert.True(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message is not received, block is requested again (to a random beacon)
	blkRequested = false
	assert.NoError(syncer.GetFailed(reachedNodeID, sentReqID))
	assert.True(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message is received from wrong validator, node waits to for the right node to respond
	blkRequested = false
	wrongNodeID := ids.ShortID{'w', 'r', 'o', 'n', 'g'}
	assert.NoError(syncer.Put(wrongNodeID, sentReqID, []byte{}))
	assert.False(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message is received with wrong reqID, node waits to for the right node to respond
	blkRequested = false
	wrongSentReqID := uint32(math.MaxUint32)
	assert.NoError(syncer.Put(reachedNodeID, wrongSentReqID, []byte{}))
	assert.False(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message carries unparsable blk, block is requested again (to a random beacon)
	blkRequested = false
	failedParseStateSyncableBlkMock := func(b []byte) (snowman.StateSyncableBlock, error) {
		return nil, fmt.Errorf("parse failed")
	}
	fullVM.ParseStateSyncableBlockF = failedParseStateSyncableBlkMock

	assert.NoError(syncer.Put(reachedNodeID, sentReqID, []byte{}))
	assert.True(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message carries the wrong blk, block is requested again (to a random beacon)
	blkRequested = false
	wrongParseStateSyncableBlkMock := func(b []byte) (snowman.StateSyncableBlock, error) {
		return &snowman.TestStateSyncableBlock{
			TestBlock: snowman.TestBlock{
				TestDecidable: choices.TestDecidable{
					IDV:     ids.ID{'w', 'r', 'o', 'n', 'g', 'I', 'D'},
					StatusV: choices.Processing,
				},
				BytesV: b,
			},
			T: t,
		}, nil
	}
	fullVM.ParseStateSyncableBlockF = wrongParseStateSyncableBlkMock

	assert.NoError(syncer.Put(reachedNodeID, sentReqID, []byte{}))
	assert.True(blkRequested)
	assert.True(reqBlkID == syncer.lastSummaryBlkID)
	assert.False(stateSyncFullyDone)

	// if Put message is received, state sync is declared done
	fullVM.ParseStateSyncableBlockF = successfulParseSyncableBlockBlkMock
	assert.NoError(syncer.Put(reachedNodeID, sentReqID, []byte{}))
	assert.True(stateSyncFullyDone)
}
