// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sampler

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedArrayInitializeOverflow(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{1, math.MaxUint64})
	assert.Error(t, err, "should have reported an overflow error")
}

func TestWeightedArrayOutOfRange(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{1})
	assert.NoError(t, err)

	err = s.StartSearch(1)
	assert.Error(t, err, "should have reported an out of range error")
}

func TestWeightedArraySingleton(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{1})
	assert.NoError(t, err)

	err = s.StartSearch(0)
	assert.NoError(t, err)

	index, ok := s.ContinueSearch()
	assert.True(t, ok, "should have found the value immediately")
	assert.Equal(t, 0, index, "should have selected the first element")
}

func TestWeightedArrayWithZero(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{0, 1})
	assert.NoError(t, err)

	err = s.StartSearch(0)
	assert.NoError(t, err)

	_, ok := s.ContinueSearch()
	assert.False(t, ok, "shouldn't have found the value immediately")

	index, ok := s.ContinueSearch()
	assert.True(t, ok, "should have found the value")
	assert.Equal(t, 1, index, "should have selected the second element")
}

func TestWeightedArrayMultiplePassesLeft(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{1, 1})
	assert.NoError(t, err)

	err = s.StartSearch(1)
	assert.NoError(t, err)

	index, ok := s.ContinueSearch()
	assert.True(t, ok, "should have found the value")
	assert.Equal(t, 1, index, "should have selected the second element")

	index, ok = s.ContinueSearch()
	assert.True(t, ok, "should have found the value")
	assert.Equal(t, 1, index, "should have selected the second element")
}

func TestWeightedArrayMultiplePassesRight(t *testing.T) {
	s := weightedArray{}
	err := s.Initialize([]uint64{1, 1, 1})
	assert.NoError(t, err)

	err = s.StartSearch(2)
	assert.NoError(t, err)

	index, ok := s.ContinueSearch()
	assert.True(t, ok, "should have found the value immediately")
	assert.Equal(t, 2, index, "should have selected the third element")

	index, ok = s.ContinueSearch()
	assert.True(t, ok, "should have already found the value")
	assert.Equal(t, 2, index, "should have selected the third element")
}
