// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package indexer

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms/proposervm/state"
	"github.com/stretchr/testify/assert"
)

var (
	genesisUnixTimestamp int64 = 1000
	genesisTimestamp           = time.Unix(genesisUnixTimestamp, 0)
)

func TestHeightBlockIndexPostFork(t *testing.T) {
	assert := assert.New(t)

	// Build a chain of wrapping blocks, representing post fork blocks
	innerBlkID := ids.Empty.Prefix(0)
	innerGenBlk := &snowman.TestBlock{
		TestDecidable: choices.TestDecidable{
			IDV:     innerBlkID,
			StatusV: choices.Accepted,
		},
		HeightV:    0,
		TimestampV: genesisTimestamp,
		BytesV:     []byte{0},
	}

	var (
		blkNumber    = uint64(10)
		lastInnerBlk = snowman.Block(innerGenBlk)
		lastProBlk   = snowman.Block(innerGenBlk)
		innerBlks    = make(map[ids.ID]snowman.Block)
		proBlks      = make(map[ids.ID]WrappingBlock)
	)
	innerBlks[innerGenBlk.ID()] = innerGenBlk

	for blkHeight := uint64(1); blkHeight <= blkNumber; blkHeight++ {
		// build inner block
		innerBlkID = ids.Empty.Prefix(blkHeight)
		innerBlk := &snowman.TestBlock{
			TestDecidable: choices.TestDecidable{
				IDV:     innerBlkID,
				StatusV: choices.Accepted,
			},
			BytesV:  []byte{uint8(blkHeight)},
			ParentV: lastInnerBlk.ID(),
			HeightV: blkHeight,
		}
		innerBlks[innerBlk.ID()] = innerBlk
		lastInnerBlk = innerBlk

		// build wrapping post fork block
		wrappingID := ids.Empty.Prefix(blkHeight + blkNumber + 1)
		postForkBlk := &TestWrappingBlock{
			TestBlock: &snowman.TestBlock{
				TestDecidable: choices.TestDecidable{
					IDV:     wrappingID,
					StatusV: choices.Accepted,
				},
				BytesV:  wrappingID[:],
				ParentV: lastProBlk.ID(),
				HeightV: lastInnerBlk.Height(),
			},
			innerBlk: innerBlk,
		}
		proBlks[postForkBlk.ID()] = postForkBlk
		lastProBlk = postForkBlk
	}

	blkSrv := &TestBlockServer{
		CantGetWrappingBlk: true,
		CantCommit:         true,

		GetWrappingBlkF: func(blkID ids.ID) (WrappingBlock, error) {
			blk, found := proBlks[blkID]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		GetInnerBlkF: func(id ids.ID) (snowman.Block, error) {
			blk, found := innerBlks[id]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		CommitF: func() error { return nil },
	}

	db := memdb.New()
	vdb := versiondb.New(db)
	storedState := state.NewHeightIndex(vdb, vdb)
	hIndex := newHeightIndexer(blkSrv,
		logging.NoLog{},
		storedState,
	)
	hIndex.commitFrequency = 0 // commit each block

	// checkpoint last accepted block and show the whole chain in reindexed
	assert.NoError(hIndex.indexState.SetCheckpoint(lastProBlk.ID()))
	assert.NoError(hIndex.RepairHeightIndex(context.Background()))
	assert.True(hIndex.IsRepaired())

	// check that height index is fully built
	loadedForkHeight, err := storedState.GetForkHeight()
	assert.NoError(err)
	assert.True(loadedForkHeight == 1)
	for height := uint64(1); height <= blkNumber; height++ {
		_, err := storedState.GetBlockIDAtHeight(height)
		assert.NoError(err)
	}
}

func TestHeightBlockIndexAcrossFork(t *testing.T) {
	assert := assert.New(t)

	// Build a chain of non-wrapping and wrapping blocks, representing pre and post fork blocks
	innerBlkID := ids.Empty.Prefix(0)
	innerGenBlk := &snowman.TestBlock{
		TestDecidable: choices.TestDecidable{
			IDV:     innerBlkID,
			StatusV: choices.Accepted,
		},
		HeightV:    0,
		TimestampV: genesisTimestamp,
		BytesV:     []byte{0},
	}

	var (
		blkNumber    = uint64(10)
		forkHeight   = blkNumber / 2
		lastInnerBlk = snowman.Block(innerGenBlk)
		lastProBlk   = snowman.Block(innerGenBlk)
		innerBlks    = make(map[ids.ID]snowman.Block)
		proBlks      = make(map[ids.ID]WrappingBlock)
	)
	innerBlks[innerGenBlk.ID()] = innerGenBlk

	for blkHeight := uint64(1); blkHeight < forkHeight; blkHeight++ {
		// build inner block
		innerBlkID = ids.Empty.Prefix(blkHeight)
		innerBlk := &snowman.TestBlock{
			TestDecidable: choices.TestDecidable{
				IDV:     innerBlkID,
				StatusV: choices.Accepted,
			},
			BytesV:  []byte{uint8(blkHeight)},
			ParentV: lastInnerBlk.ID(),
			HeightV: blkHeight,
		}
		innerBlks[innerBlk.ID()] = innerBlk
		lastInnerBlk = innerBlk
	}

	for blkHeight := forkHeight; blkHeight <= blkNumber; blkHeight++ {
		// build inner block
		innerBlkID = ids.Empty.Prefix(blkHeight)
		innerBlk := &snowman.TestBlock{
			TestDecidable: choices.TestDecidable{
				IDV:     innerBlkID,
				StatusV: choices.Accepted,
			},
			BytesV:  []byte{uint8(blkHeight)},
			ParentV: lastInnerBlk.ID(),
			HeightV: blkHeight,
		}
		innerBlks[innerBlk.ID()] = innerBlk
		lastInnerBlk = innerBlk

		// build wrapping post fork block
		wrappingID := ids.Empty.Prefix(blkHeight + blkNumber + 1)
		postForkBlk := &TestWrappingBlock{
			TestBlock: &snowman.TestBlock{
				TestDecidable: choices.TestDecidable{
					IDV:     wrappingID,
					StatusV: choices.Accepted,
				},
				BytesV:  wrappingID[:],
				ParentV: lastProBlk.ID(),
				HeightV: lastInnerBlk.Height(),
			},
			innerBlk: innerBlk,
		}
		proBlks[postForkBlk.ID()] = postForkBlk
		lastProBlk = postForkBlk
	}

	blkSrv := &TestBlockServer{
		CantGetWrappingBlk: true,
		CantCommit:         true,

		GetWrappingBlkF: func(blkID ids.ID) (WrappingBlock, error) {
			blk, found := proBlks[blkID]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		GetInnerBlkF: func(id ids.ID) (snowman.Block, error) {
			blk, found := innerBlks[id]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		CommitF: func() error { return nil },
	}

	db := memdb.New()
	vdb := versiondb.New(db)
	storedState := state.NewHeightIndex(vdb, vdb)
	hIndex := newHeightIndexer(blkSrv,
		logging.NoLog{},
		storedState,
	)
	hIndex.commitFrequency = 0 // commit each block

	// checkpoint last accepted block and show the whole chain in reindexed
	assert.NoError(hIndex.indexState.SetCheckpoint(lastProBlk.ID()))
	assert.NoError(hIndex.RepairHeightIndex(context.Background()))
	assert.True(hIndex.IsRepaired())

	// check that height index is fully built
	loadedForkHeight, err := storedState.GetForkHeight()
	assert.NoError(err)
	assert.True(loadedForkHeight == forkHeight)
	for height := uint64(0); height < forkHeight; height++ {
		_, err := storedState.GetBlockIDAtHeight(height)
		assert.Error(err, database.ErrNotFound)
	}
	for height := forkHeight; height <= blkNumber; height++ {
		_, err := storedState.GetBlockIDAtHeight(height)
		assert.NoError(err)
	}
}

func TestHeightBlockIndexResumeFromCheckPoint(t *testing.T) {
	assert := assert.New(t)

	// Build a chain of non-wrapping and wrapping blocks, representing pre and post fork blocks
	innerBlkID := ids.Empty.Prefix(0)
	innerGenBlk := &snowman.TestBlock{
		TestDecidable: choices.TestDecidable{
			IDV:     innerBlkID,
			StatusV: choices.Accepted,
		},
		HeightV:    0,
		TimestampV: genesisTimestamp,
		BytesV:     []byte{0},
	}

	var (
		blkNumber  = uint64(10)
		forkHeight = blkNumber / 2

		lastInnerBlk = snowman.Block(innerGenBlk)
		lastProBlk   = snowman.Block(innerGenBlk)

		innerBlks = make(map[ids.ID]snowman.Block)
		proBlks   = make(map[ids.ID]WrappingBlock)
	)
	innerBlks[innerGenBlk.ID()] = innerGenBlk

	for blkHeight := uint64(1); blkHeight < forkHeight; blkHeight++ {
		// build inner block
		innerBlkID = ids.Empty.Prefix(blkHeight)
		innerBlk := &snowman.TestBlock{
			TestDecidable: choices.TestDecidable{
				IDV:     innerBlkID,
				StatusV: choices.Accepted,
			},
			BytesV:  []byte{uint8(blkHeight)},
			ParentV: lastInnerBlk.ID(),
			HeightV: blkHeight,
		}
		innerBlks[innerBlk.ID()] = innerBlk
		lastInnerBlk = innerBlk
	}

	for blkHeight := forkHeight; blkHeight <= blkNumber; blkHeight++ {
		// build inner block
		innerBlkID = ids.Empty.Prefix(blkHeight)
		innerBlk := &snowman.TestBlock{
			TestDecidable: choices.TestDecidable{
				IDV:     innerBlkID,
				StatusV: choices.Accepted,
			},
			BytesV:  []byte{uint8(blkHeight)},
			ParentV: lastInnerBlk.ID(),
			HeightV: blkHeight,
		}
		innerBlks[innerBlk.ID()] = innerBlk
		lastInnerBlk = innerBlk

		// build wrapping post fork block
		wrappingID := ids.Empty.Prefix(blkHeight + blkNumber + 1)
		postForkBlk := &TestWrappingBlock{
			TestBlock: &snowman.TestBlock{
				TestDecidable: choices.TestDecidable{
					IDV:     wrappingID,
					StatusV: choices.Accepted,
				},
				BytesV:  wrappingID[:],
				ParentV: lastProBlk.ID(),
				HeightV: lastInnerBlk.Height(),
			},
			innerBlk: innerBlk,
		}
		proBlks[postForkBlk.ID()] = postForkBlk
		lastProBlk = postForkBlk
	}

	blkSrv := &TestBlockServer{
		CantGetWrappingBlk: true,
		CantCommit:         true,

		GetWrappingBlkF: func(blkID ids.ID) (WrappingBlock, error) {
			blk, found := proBlks[blkID]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		GetInnerBlkF: func(id ids.ID) (snowman.Block, error) {
			blk, found := innerBlks[id]
			if !found {
				return nil, database.ErrNotFound
			}
			return blk, nil
		},
		CommitF: func() error { return nil },
	}

	db := memdb.New()
	vdb := versiondb.New(db)
	storedState := state.NewHeightIndex(vdb, vdb)
	hIndex := newHeightIndexer(blkSrv,
		logging.NoLog{},
		storedState,
	)
	hIndex.commitFrequency = 0 // commit each block

	// pick a random block in the chain and checkpoint it;...
	rndPostForkHeight := rand.Intn(int(blkNumber-forkHeight)) + int(forkHeight) // #nosec G404
	var checkpointBlk WrappingBlock
	for _, blk := range proBlks {
		if blk.Height() != uint64(rndPostForkHeight) {
			continue // not the blk we are looking for
		}

		checkpointBlk = blk
		assert.NoError(hIndex.indexState.SetCheckpoint(checkpointBlk.ID()))
		break
	}

	// perform repair and show index is built
	assert.NoError(hIndex.RepairHeightIndex(context.Background()))
	assert.True(hIndex.IsRepaired())

	// check that height index is fully built
	loadedForkHeight, err := storedState.GetForkHeight()
	assert.NoError(err)
	assert.True(loadedForkHeight == forkHeight)
	for height := forkHeight; height <= checkpointBlk.Height(); height++ {
		_, err := storedState.GetBlockIDAtHeight(height)
		assert.NoError(err)
	}
}
