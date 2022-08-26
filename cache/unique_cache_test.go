// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cache

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

type evictable[T comparable] struct {
	id      T
	evicted int
}

func (e *evictable[T]) Key() T { return e.id }
func (e *evictable[T]) Evict() { e.evicted++ }

func TestEvictableLRU(t *testing.T) {
	cache := EvictableLRU[ids.ID, int]{}

	expectedValue1 := &evictable[ids.ID]{id: ids.ID{1}}
	if returnedValue := cache.Deduplicate(expectedValue1).(*evictable[ids.ID]); returnedValue != expectedValue1 {
		t.Fatalf("Returned unknown value")
	} else if expectedValue1.evicted != 0 {
		t.Fatalf("Value was evicted unexpectedly")
	} else if returnedValue := cache.Deduplicate(expectedValue1).(*evictable[ids.ID]); returnedValue != expectedValue1 {
		t.Fatalf("Returned unknown value")
	} else if expectedValue1.evicted != 0 {
		t.Fatalf("Value was evicted unexpectedly")
	}

	expectedValue2 := &evictable[ids.ID]{id: ids.ID{2}}
	returnedValue := cache.Deduplicate(expectedValue2).(*evictable[ids.ID])
	switch {
	case returnedValue != expectedValue2:
		t.Fatalf("Returned unknown value")
	case expectedValue1.evicted != 1:
		t.Fatalf("Value should have been evicted")
	case expectedValue2.evicted != 0:
		t.Fatalf("Value was evicted unexpectedly")
	}

	cache.Size = 2

	expectedValue3 := &evictable[ids.ID]{id: ids.ID{2}}
	returnedValue = cache.Deduplicate(expectedValue3).(*evictable[ids.ID])
	switch {
	case returnedValue != expectedValue2:
		t.Fatalf("Returned unknown value")
	case expectedValue1.evicted != 1:
		t.Fatalf("Value should have been evicted")
	case expectedValue2.evicted != 0:
		t.Fatalf("Value was evicted unexpectedly")
	}

	cache.Flush()
	switch {
	case expectedValue1.evicted != 1:
		t.Fatalf("Value should have been evicted")
	case expectedValue2.evicted != 1:
		t.Fatalf("Value should have been evicted")
	case expectedValue3.evicted != 0:
		t.Fatalf("Value was evicted unexpectedly")
	}
}
