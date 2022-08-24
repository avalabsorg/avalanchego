// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validators

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/sampler"
	"github.com/ava-labs/avalanchego/utils/set"

	safemath "github.com/ava-labs/avalanchego/utils/math"
)

const (
	// If, when the validator set is reset, cap(set)/len(set) > MaxExcessCapacityFactor,
	// the underlying arrays' capacities will be reduced by a factor of capacityReductionFactor.
	// Higher value for maxExcessCapacityFactor --> less aggressive array downsizing --> less memory allocations
	// but more unnecessary data in the underlying array that can't be garbage collected.
	// Higher value for capacityReductionFactor --> more aggressive array downsizing --> more memory allocations
	// but less unnecessary data in the underlying array that can't be garbage collected.
	maxExcessCapacityFactor = 4
	capacityReductionFactor = 2
)

var _ Set = &vdrSet{}

// Set of validators that can be sampled
type Set interface {
	fmt.Stringer
	PrefixedString(string) string

	// Set removes all the current validators and adds all the provided
	// validators to the set.
	Set([]Validator) error

	// AddWeight to a staker.
	AddWeight(ids.NodeID, uint64) error

	// GetWeight retrieves the validator weight from the set.
	GetWeight(ids.NodeID) (uint64, bool)

	// SubsetWeight returns the sum of the weights of the validators.
	SubsetWeight(set.Set[ids.NodeID]) (uint64, error)

	// RemoveWeight from a staker.
	RemoveWeight(ids.NodeID, uint64) error

	// Contains returns true if there is a validator with the specified ID
	// currently in the set.
	Contains(ids.NodeID) bool

	// Len returns the number of validators currently in the set.
	Len() int

	// List all the validators in this group
	List() []Validator

	// Weight returns the cumulative weight of all validators in the set.
	Weight() uint64

	// Sample returns a collection of validators, potentially with duplicates.
	// If sampling the requested size isn't possible, an error will be returned.
	Sample(size int) ([]Validator, error)

	// MaskValidator hides the named validator from future samplings
	MaskValidator(ids.NodeID) error

	// When a validator's weight changes, or a validator is added/removed,
	// this listener is called.
	RegisterCallbackListener(SetCallbackListener)

	RevealValidator(ids.NodeID) error
}

type SetCallbackListener interface {
	OnValidatorAdded(validatorID ids.NodeID, weight uint64)
	OnValidatorRemoved(validatorID ids.NodeID, weight uint64)
	OnValidatorWeightChanged(validatorID ids.NodeID, oldWeight, newWeight uint64)
}

// NewSet returns a new, empty set of validators.
func NewSet() Set {
	return &vdrSet{
		vdrMap:  make(map[ids.NodeID]int),
		sampler: sampler.NewWeightedWithoutReplacement(),
	}
}

// NewBestSet returns a new, empty set of validators.
func NewBestSet(expectedSampleSize int) Set {
	return &vdrSet{
		vdrMap:  make(map[ids.NodeID]int),
		sampler: sampler.NewBestWeightedWithoutReplacement(expectedSampleSize),
	}
}

// vdrSet of validators. Validator function results are cached. Therefore, to
// update a validators weight, one should ensure to call add with the updated
// validator.
type vdrSet struct {
	initialized       bool
	lock              sync.RWMutex
	vdrMap            map[ids.NodeID]int
	vdrSlice          []*validator
	vdrWeights        []uint64
	vdrMaskedWeights  []uint64
	sampler           sampler.WeightedWithoutReplacement
	totalWeight       uint64
	maskedVdrs        set.Set[ids.NodeID]
	callbackListeners []SetCallbackListener
}

func (s *vdrSet) Set(vdrs []Validator) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.set(vdrs)
}

func (s *vdrSet) set(vdrs []Validator) error {
	// find all the nodes that are going to be added or have their weight changed
	nodesInResultSet := set.NewSet[ids.NodeID](len(vdrs))
	for _, vdr := range vdrs {
		vdrID := vdr.ID()
		if nodesInResultSet.Contains(vdrID) {
			continue
		}
		nodesInResultSet.Add(vdrID)

		newWeight := vdr.Weight()
		index, contains := s.vdrMap[vdrID]
		if !contains {
			s.callValidatorAddedCallbacks(vdrID, newWeight)
			continue
		}

		existingWeight := s.vdrWeights[index]
		if existingWeight != newWeight {
			s.callWeightChangeCallbacks(vdrID, existingWeight, newWeight)
		}
	}

	// find all nodes that are going to be removed
	for nodeID, index := range s.vdrMap {
		if !nodesInResultSet.Contains(nodeID) {
			s.callValidatorRemovedCallbacks(nodeID, s.vdrWeights[index])
		}
	}

	lenVdrs := len(vdrs)
	// If the underlying arrays are much larger than necessary, resize them to
	// allow garbage collection of unused memory
	if cap(s.vdrSlice) > len(s.vdrSlice)*maxExcessCapacityFactor {
		newCap := cap(s.vdrSlice) / capacityReductionFactor
		if newCap < lenVdrs {
			newCap = lenVdrs
		}
		s.vdrSlice = make([]*validator, 0, newCap)
		s.vdrWeights = make([]uint64, 0, newCap)
		s.vdrMaskedWeights = make([]uint64, 0, newCap)
	} else {
		s.vdrSlice = s.vdrSlice[:0]
		s.vdrWeights = s.vdrWeights[:0]
		s.vdrMaskedWeights = s.vdrMaskedWeights[:0]
	}
	s.vdrMap = make(map[ids.NodeID]int, lenVdrs)
	s.totalWeight = 0
	s.initialized = false

	for _, vdr := range vdrs {
		vdrID := vdr.ID()
		if s.contains(vdrID) {
			continue
		}
		w := vdr.Weight()
		if w == 0 {
			continue // This validator would never be sampled anyway
		}

		i := len(s.vdrSlice)
		s.vdrMap[vdrID] = i
		s.vdrSlice = append(s.vdrSlice, &validator{
			nodeID: vdr.ID(),
			weight: vdr.Weight(),
		})
		s.vdrWeights = append(s.vdrWeights, w)
		s.vdrMaskedWeights = append(s.vdrMaskedWeights, 0)

		if s.maskedVdrs.Contains(vdrID) {
			continue
		}
		s.vdrMaskedWeights[len(s.vdrMaskedWeights)-1] = w

		newTotalWeight, err := safemath.Add64(s.totalWeight, w)
		if err != nil {
			return err
		}
		s.totalWeight = newTotalWeight
	}
	return nil
}

func (s *vdrSet) AddWeight(vdrID ids.NodeID, weight uint64) error {
	if weight == 0 {
		return nil // This validator would never be sampled anyway
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.addWeight(vdrID, weight)
}

func (s *vdrSet) addWeight(vdrID ids.NodeID, weight uint64) error {
	var vdr *validator
	i, nodeExists := s.vdrMap[vdrID]
	if !nodeExists {
		vdr = &validator{
			nodeID: vdrID,
		}
		i = len(s.vdrSlice)
		s.vdrSlice = append(s.vdrSlice, vdr)
		s.vdrWeights = append(s.vdrWeights, 0)
		s.vdrMaskedWeights = append(s.vdrMaskedWeights, 0)
		s.vdrMap[vdrID] = i
		s.callValidatorAddedCallbacks(vdrID, weight)
	} else {
		vdr = s.vdrSlice[i]
	}

	oldWeight := s.vdrWeights[i]
	s.vdrWeights[i] += weight
	vdr.addWeight(weight)

	if nodeExists {
		s.callWeightChangeCallbacks(vdrID, oldWeight, vdr.weight)
	}

	if s.maskedVdrs.Contains(vdrID) {
		return nil
	}
	s.vdrMaskedWeights[i] += weight

	newTotalWeight, err := safemath.Add64(s.totalWeight, weight)
	if err != nil {
		return nil
	}
	s.totalWeight = newTotalWeight
	s.initialized = false
	return nil
}

func (s *vdrSet) GetWeight(vdrID ids.NodeID) (uint64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.getWeight(vdrID)
}

func (s *vdrSet) getWeight(vdrID ids.NodeID) (uint64, bool) {
	if index, ok := s.vdrMap[vdrID]; ok {
		return s.vdrMaskedWeights[index], true
	}
	return 0, false
}

func (s *vdrSet) SubsetWeight(subset set.Set[ids.NodeID]) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	totalWeight := uint64(0)
	for vdrID := range subset {
		weight, ok := s.getWeight(vdrID)
		if !ok {
			continue
		}
		newWeight, err := safemath.Add64(totalWeight, weight)
		if err != nil {
			return 0, err
		}
		totalWeight = newWeight
	}
	return totalWeight, nil
}

func (s *vdrSet) RemoveWeight(vdrID ids.NodeID, weight uint64) error {
	if weight == 0 {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.removeWeight(vdrID, weight)
}

func (s *vdrSet) removeWeight(vdrID ids.NodeID, weight uint64) error {
	i, ok := s.vdrMap[vdrID]
	if !ok {
		return nil
	}

	// Validator exists
	vdr := s.vdrSlice[i]

	oldWeight := s.vdrWeights[i]
	weight = safemath.Min64(oldWeight, weight)
	s.vdrWeights[i] -= weight
	vdr.removeWeight(weight)
	if !s.maskedVdrs.Contains(vdrID) {
		s.totalWeight -= weight
		s.vdrMaskedWeights[i] -= weight
	}

	if vdr.Weight() == 0 {
		s.callValidatorRemovedCallbacks(vdrID, oldWeight)
		if err := s.remove(vdrID); err != nil {
			return err
		}
	} else {
		s.callWeightChangeCallbacks(vdrID, oldWeight, vdr.weight)
	}
	s.initialized = false
	return nil
}

func (s *vdrSet) Get(vdrID ids.NodeID) (Validator, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.get(vdrID)
}

func (s *vdrSet) get(vdrID ids.NodeID) (Validator, bool) {
	index, ok := s.vdrMap[vdrID]
	if !ok {
		return nil, false
	}
	return s.vdrSlice[index], true
}

func (s *vdrSet) remove(vdrID ids.NodeID) error {
	// Get the element to remove
	i, contains := s.vdrMap[vdrID]
	if !contains {
		return nil
	}

	// Get the last element
	e := len(s.vdrSlice) - 1
	eVdr := s.vdrSlice[e]

	// Move e -> i
	iElem := s.vdrSlice[i]
	s.vdrMap[eVdr.ID()] = i
	s.vdrSlice[i] = eVdr
	s.vdrWeights[i] = s.vdrWeights[e]
	s.vdrMaskedWeights[i] = s.vdrMaskedWeights[e]

	// Remove i
	delete(s.vdrMap, vdrID)
	s.vdrSlice = s.vdrSlice[:e]
	s.vdrWeights = s.vdrWeights[:e]
	s.vdrMaskedWeights = s.vdrMaskedWeights[:e]

	if !s.maskedVdrs.Contains(vdrID) {
		newTotalWeight, err := safemath.Sub64(s.totalWeight, iElem.Weight())
		if err != nil {
			return err
		}
		s.totalWeight = newTotalWeight
	}
	s.initialized = false
	return nil
}

func (s *vdrSet) Contains(vdrID ids.NodeID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.contains(vdrID)
}

func (s *vdrSet) contains(vdrID ids.NodeID) bool {
	_, contains := s.vdrMap[vdrID]
	return contains
}

func (s *vdrSet) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.len()
}

func (s *vdrSet) len() int { return len(s.vdrSlice) }

func (s *vdrSet) List() []Validator {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.list()
}

func (s *vdrSet) list() []Validator {
	list := make([]Validator, len(s.vdrSlice))
	for i, vdr := range s.vdrSlice {
		list[i] = vdr
	}
	return list
}

func (s *vdrSet) Sample(size int) ([]Validator, error) {
	if size == 0 {
		return nil, nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.sample(size)
}

func (s *vdrSet) sample(size int) ([]Validator, error) {
	if !s.initialized {
		if err := s.sampler.Initialize(s.vdrMaskedWeights); err != nil {
			return nil, err
		}
		s.initialized = true
	}
	indices, err := s.sampler.Sample(size)
	if err != nil {
		return nil, err
	}

	list := make([]Validator, size)
	for i, index := range indices {
		list[i] = s.vdrSlice[index]
	}
	return list, nil
}

func (s *vdrSet) Weight() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.totalWeight
}

func (s *vdrSet) String() string {
	return s.PrefixedString("")
}

func (s *vdrSet) PrefixedString(prefix string) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.prefixedString(prefix)
}

func (s *vdrSet) prefixedString(prefix string) string {
	sb := strings.Builder{}

	totalWeight := uint64(0)
	for _, weight := range s.vdrWeights {
		totalWeight += weight
	}

	sb.WriteString(fmt.Sprintf("Validator Set: (Size = %d, SampleableWeight = %d, Weight = %d)",
		len(s.vdrSlice),
		s.totalWeight,
		totalWeight,
	))
	format := fmt.Sprintf("\n%s    Validator[%s]: %%33s, %%d/%%d", prefix, formatting.IntFormat(len(s.vdrSlice)-1))
	for i, vdr := range s.vdrSlice {
		sb.WriteString(fmt.Sprintf(format,
			i,
			vdr.ID(),
			s.vdrMaskedWeights[i],
			vdr.Weight()))
	}

	return sb.String()
}

func (s *vdrSet) MaskValidator(vdrID ids.NodeID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.maskValidator(vdrID)
}

func (s *vdrSet) maskValidator(vdrID ids.NodeID) error {
	if s.maskedVdrs.Contains(vdrID) {
		return nil
	}

	s.maskedVdrs.Add(vdrID)

	// Get the element to mask
	i, contains := s.vdrMap[vdrID]
	if !contains {
		return nil
	}

	s.vdrMaskedWeights[i] = 0
	s.totalWeight -= s.vdrWeights[i]
	s.initialized = false

	return nil
}

func (s *vdrSet) RevealValidator(vdrID ids.NodeID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.revealValidator(vdrID)
}

func (s *vdrSet) revealValidator(vdrID ids.NodeID) error {
	if !s.maskedVdrs.Contains(vdrID) {
		return nil
	}

	s.maskedVdrs.Remove(vdrID)

	// Get the element to reveal
	i, contains := s.vdrMap[vdrID]
	if !contains {
		return nil
	}

	weight := s.vdrWeights[i]
	s.vdrMaskedWeights[i] = weight
	newTotalWeight, err := safemath.Add64(s.totalWeight, weight)
	if err != nil {
		return err
	}
	s.totalWeight = newTotalWeight
	s.initialized = false

	return nil
}

func (s *vdrSet) RegisterCallbackListener(callbackListener SetCallbackListener) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.callbackListeners = append(s.callbackListeners, callbackListener)
	for node, index := range s.vdrMap {
		callbackListener.OnValidatorAdded(node, s.vdrWeights[index])
	}
}

// Assumes [s.lock] is held
func (s *vdrSet) callWeightChangeCallbacks(node ids.NodeID, oldWeight, newWeight uint64) {
	for _, callbackListener := range s.callbackListeners {
		callbackListener.OnValidatorWeightChanged(node, oldWeight, newWeight)
	}
}

// Assumes [s.lock] is held
func (s *vdrSet) callValidatorAddedCallbacks(node ids.NodeID, weight uint64) {
	for _, callbackListener := range s.callbackListeners {
		callbackListener.OnValidatorAdded(node, weight)
	}
}

// Assumes [s.lock] is held
func (s *vdrSet) callValidatorRemovedCallbacks(node ids.NodeID, weight uint64) {
	for _, callbackListener := range s.callbackListeners {
		callbackListener.OnValidatorRemoved(node, weight)
	}
}
