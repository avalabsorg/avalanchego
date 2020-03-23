// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"errors"

	"github.com/ava-labs/gecko/cache"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/vms/components/ava"
)

var (
	errCacheTypeMismatch = errors.New("type returned from cache doesn't match the expected type")
)

func uniqueID(id ids.ID, prefix uint64, cacher cache.Cacher) ids.ID {
	if cachedIDIntf, found := cacher.Get(id); found {
		return cachedIDIntf.(ids.ID)
	}
	uID := id.Prefix(prefix)
	cacher.Put(id, uID)
	return uID
}

// state is a thin wrapper around a database to provide, caching, serialization,
// and de-serialization.
type state struct{ ava.State }

// Tx attempts to load a transaction from storage.
func (s *state) Tx(id ids.ID) (*Tx, error) {
	if txIntf, found := s.Cache.Get(id); found {
		if tx, ok := txIntf.(*Tx); ok {
			return tx, nil
		}
		return nil, errCacheTypeMismatch
	}

	bytes, err := s.DB.Get(id.Bytes())
	if err != nil {
		return nil, err
	}

	// The key was in the database
	tx := &Tx{}
	if err := s.Codec.Unmarshal(bytes, tx); err != nil {
		return nil, err
	}
	tx.Initialize(bytes)

	s.Cache.Put(id, tx)
	return tx, nil
}

// SetTx saves the provided transaction to storage.
func (s *state) SetTx(id ids.ID, tx *Tx) error {
	if tx == nil {
		s.Cache.Evict(id)
		return s.DB.Delete(id.Bytes())
	}

	s.Cache.Put(id, tx)
	return s.DB.Put(id.Bytes(), tx.Bytes())
}

// IDs returns a slice of IDs from storage
func (s *state) IDs(id ids.ID) ([]ids.ID, error) {
	if idsIntf, found := s.Cache.Get(id); found {
		if idSlice, ok := idsIntf.([]ids.ID); ok {
			return idSlice, nil
		}
		return nil, errCacheTypeMismatch
	}

	bytes, err := s.DB.Get(id.Bytes())
	if err != nil {
		return nil, err
	}

	idSlice := []ids.ID(nil)
	if err := s.Codec.Unmarshal(bytes, &idSlice); err != nil {
		return nil, err
	}

	s.Cache.Put(id, idSlice)
	return idSlice, nil
}

// SetIDs saves a slice of IDs to the database.
func (s *state) SetIDs(id ids.ID, idSlice []ids.ID) error {
	if len(idSlice) == 0 {
		s.Cache.Evict(id)
		return s.DB.Delete(id.Bytes())
	}

	s.Cache.Put(id, idSlice)

	bytes, err := s.Codec.Marshal(idSlice)
	if err != nil {
		return err
	}

	return s.DB.Put(id.Bytes(), bytes)
}
