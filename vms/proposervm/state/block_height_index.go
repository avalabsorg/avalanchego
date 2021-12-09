// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"encoding/binary"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

const (
	cacheSize = 8192
)

var (
	_ AcceptedPostForkBlockHeightIndex = &innerBlocksMapping{}

	heightPrefix  = []byte("heightkey")
	preForkPrefix = []byte("preForkKey")
	resumePrefix  = []byte("blockToResumeFrom")
)

// AcceptedPostForkBlockHeightIndex contains mapping of blockHeights to accepted proposer block IDs.
// Only accepted blocks are indexed; moreover only post-fork blocks are indexed.
type AcceptedPostForkBlockHeightIndex interface {
	SetBlkIDByHeight(height uint64, blkID ids.ID) (int, error)
	GetBlkIDByHeight(height uint64) (ids.ID, error)
	DeleteBlkIDByHeight(height uint64) error

	SetLatestPreForkHeight(height uint64) error
	GetLatestPreForkHeight() (uint64, error)
	DeleteLatestPreForkHeight() error

	SetBlockToResumeFrom(blkID ids.ID) error
	GetBlockToResumeFrom() (ids.ID, error)
	DeleteBlockToResumeFrom() error

	clearCache() // useful in testing
}

type innerBlocksMapping struct {
	// Caches block height -> proposerVMBlockID. If the proposerVMBlockID is nil,
	// the height is not in storage.
	cache cache.Cacher

	db database.Database
}

func NewBlockHeightIndex(db database.Database) AcceptedPostForkBlockHeightIndex {
	return &innerBlocksMapping{
		cache: &cache.LRU{Size: cacheSize},
		db:    db,
	}
}

func (ibm *innerBlocksMapping) SetBlkIDByHeight(height uint64, blkID ids.ID) (int, error) {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)
	key := make([]byte, len(heightPrefix)+len(heightBytes))
	copy(key, heightPrefix)
	key = append(key, heightBytes...)

	ibm.cache.Put(string(key), blkID)
	return len(key) + len(blkID), ibm.db.Put(key, blkID[:])
}

func (ibm *innerBlocksMapping) GetBlkIDByHeight(height uint64) (ids.ID, error) {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)
	key := make([]byte, len(heightPrefix)+len(heightBytes))
	copy(key, heightPrefix)
	key = append(key, heightBytes...)

	if blkIDIntf, found := ibm.cache.Get(string(key)); found {
		if blkIDIntf == nil {
			return ids.Empty, database.ErrNotFound
		}

		res, _ := blkIDIntf.(ids.ID)
		return res, nil
	}

	bytes, err := ibm.db.Get(key)
	switch err {
	case nil:
		var res ids.ID
		copy(res[:], bytes)
		return res, nil

	case database.ErrNotFound:
		ibm.cache.Put(string(key), nil)
		return ids.Empty, database.ErrNotFound

	default:
		return ids.Empty, err
	}
}

func (ibm *innerBlocksMapping) DeleteBlkIDByHeight(height uint64) error {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)
	key := make([]byte, len(heightPrefix)+len(heightBytes))
	copy(key, heightPrefix)
	key = append(key, heightBytes...)

	ibm.cache.Put(string(key), nil)
	return ibm.db.Delete(key)
}

func (ibm *innerBlocksMapping) SetLatestPreForkHeight(height uint64) error {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)

	ibm.cache.Put(string(preForkPrefix), heightBytes)
	return ibm.db.Put(preForkPrefix, heightBytes)
}

func (ibm *innerBlocksMapping) GetLatestPreForkHeight() (uint64, error) {
	key := preForkPrefix
	if blkIDIntf, found := ibm.cache.Get(string(key)); found {
		if blkIDIntf == nil {
			return 0, database.ErrNotFound
		}

		heightBytes, _ := blkIDIntf.([]byte)
		res := binary.BigEndian.Uint64(heightBytes)
		return res, nil
	}

	bytes, err := ibm.db.Get(key)
	switch err {
	case nil:
		res := binary.BigEndian.Uint64(bytes)
		return res, nil

	case database.ErrNotFound:
		ibm.cache.Put(string(key), nil)
		return 0, database.ErrNotFound

	default:
		return 0, err
	}
}

func (ibm *innerBlocksMapping) DeleteLatestPreForkHeight() error {
	key := preForkPrefix
	ibm.cache.Put(string(key), nil)
	return ibm.db.Delete(key)
}

func (ibm *innerBlocksMapping) SetBlockToResumeFrom(blkID ids.ID) error {
	key := resumePrefix
	ibm.cache.Put(string(key), blkID)
	return ibm.db.Put(key, blkID[:])
}

func (ibm *innerBlocksMapping) GetBlockToResumeFrom() (ids.ID, error) {
	key := resumePrefix
	if blkIDIntf, found := ibm.cache.Get(string(key)); found {
		if blkIDIntf == nil {
			return ids.Empty, database.ErrNotFound
		}

		res, _ := blkIDIntf.(ids.ID)
		return res, nil
	}

	bytes, err := ibm.db.Get(key)
	switch err {
	case nil:
		var res ids.ID
		copy(res[:], bytes)
		return res, nil

	case database.ErrNotFound:
		ibm.cache.Put(string(key), nil)
		return ids.Empty, database.ErrNotFound

	default:
		return ids.Empty, err
	}
}

func (ibm *innerBlocksMapping) DeleteBlockToResumeFrom() error {
	key := resumePrefix
	ibm.cache.Put(string(key), nil)
	return ibm.db.Delete(key)
}

func (ibm *innerBlocksMapping) clearCache() {
	ibm.cache.Flush()
}
