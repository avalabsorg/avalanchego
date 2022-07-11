// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/google/btree"
)

type Validators interface {
	CurrentValidators
	PendingValidators
}

type CurrentValidators interface {
	GetCurrentStaker(subnetID ids.ID, nodeID ids.NodeID) (*Staker, error)
	PutCurrentStaker(staker *Staker)
	DeleteCurrentStaker(staker *Staker)

	GetCurrentDelegatorIterator(subnetID ids.ID, nodeID ids.NodeID) (StakerIterator, error)
	PutCurrentDelegator(staker *Staker)
	DeleteCurrentDelegator(staker *Staker)

	// GetCurrentStakerIterator returns the stakers in the validator set sorted
	// in order of their future removal.
	GetCurrentStakerIterator() (StakerIterator, error)
}

type PendingValidators interface {
	GetPendingStaker(subnetID ids.ID, nodeID ids.NodeID) (*Staker, error)
	PutPendingStaker(staker *Staker)
	DeletePendingStaker(staker *Staker)

	GetPendingDelegatorIterator(subnetID ids.ID, nodeID ids.NodeID) (StakerIterator, error)
	PutPendingDelegator(staker *Staker)
	DeletePendingDelegator(staker *Staker)

	// GetPendingStakerIterator returns the stakers in the validator set sorted
	// in order of their future removal.
	GetPendingStakerIterator() (StakerIterator, error)
}

type baseValidators struct {
	validators     map[ids.ID]map[ids.NodeID]*baseValidator
	stakers        *btree.BTree
	validatorDiffs map[ids.ID]map[ids.NodeID]*diffValidator
}

func newBaseValidators() *baseValidators {
	return &baseValidators{
		validators:     make(map[ids.ID]map[ids.NodeID]*baseValidator),
		stakers:        btree.New(defaultTreeDegree),
		validatorDiffs: make(map[ids.ID]map[ids.NodeID]*diffValidator),
	}
}

func (v *baseValidators) GetStaker(subnetID ids.ID, nodeID ids.NodeID) (*Staker, error) {
	subnetValidators, ok := v.validators[subnetID]
	if !ok {
		return nil, database.ErrNotFound
	}
	validator, ok := subnetValidators[nodeID]
	if !ok {
		return nil, database.ErrNotFound
	}
	if validator.staker != nil {
		return nil, database.ErrNotFound
	}
	return validator.staker, nil
}

func (v *baseValidators) PutStaker(staker *Staker) {
	validator := v.getOrCreateValidator(staker.SubnetID, staker.NodeID)
	validator.staker = staker

	validatorDiff := v.getOrCreateValidatorDiff(staker.SubnetID, staker.NodeID)
	validatorDiff.stakerModified = true
	validatorDiff.stakerDeleted = false
	validatorDiff.staker = staker

	v.stakers.ReplaceOrInsert(staker)
}

func (v *baseValidators) DeleteStaker(staker *Staker) {
	validator := v.getOrCreateValidator(staker.SubnetID, staker.NodeID)
	validator.staker = nil
	v.pruneValidator(staker.SubnetID, staker.NodeID)

	validatorDiff := v.getOrCreateValidatorDiff(staker.SubnetID, staker.NodeID)
	validatorDiff.stakerModified = true
	validatorDiff.stakerDeleted = true
	validatorDiff.staker = staker

	v.stakers.Delete(staker)
}

func (v *baseValidators) GetDelegatorIterator(subnetID ids.ID, nodeID ids.NodeID) StakerIterator {
	subnetValidators, ok := v.validators[subnetID]
	if !ok {
		return EmptyIterator
	}
	validator, ok := subnetValidators[nodeID]
	if !ok {
		return EmptyIterator
	}
	return NewTreeIterator(validator.delegators)
}

func (v *baseValidators) PutDelegator(staker *Staker) {
	validator := v.getOrCreateValidator(staker.SubnetID, staker.NodeID)
	if validator.delegators == nil {
		validator.delegators = btree.New(defaultTreeDegree)
	}
	validator.delegators.ReplaceOrInsert(staker)

	validatorDiff := v.getOrCreateValidatorDiff(staker.SubnetID, staker.NodeID)
	if validatorDiff.addedDelegators == nil {
		validatorDiff.addedDelegators = btree.New(defaultTreeDegree)
	}
	validatorDiff.addedDelegators.ReplaceOrInsert(staker)

	v.stakers.ReplaceOrInsert(staker)
}

func (v *baseValidators) DeleteDelegator(staker *Staker) {
	validator := v.getOrCreateValidator(staker.SubnetID, staker.NodeID)
	if validator.delegators != nil {
		validator.delegators.Delete(staker)
	}
	v.pruneValidator(staker.SubnetID, staker.NodeID)

	validatorDiff := v.getOrCreateValidatorDiff(staker.SubnetID, staker.NodeID)
	if validatorDiff.deletedDelegators == nil {
		validatorDiff.deletedDelegators = make(map[ids.ID]*Staker)
	}
	validatorDiff.deletedDelegators[staker.TxID] = staker

	v.stakers.Delete(staker)
}

func (v *baseValidators) GetStakerIterator() StakerIterator {
	return NewTreeIterator(v.stakers)
}

func (v *baseValidators) getOrCreateValidator(subnetID ids.ID, nodeID ids.NodeID) *baseValidator {
	subnetValidators, ok := v.validators[subnetID]
	if !ok {
		subnetValidators = make(map[ids.NodeID]*baseValidator)
		v.validators[subnetID] = subnetValidators
	}
	validator, ok := subnetValidators[nodeID]
	if !ok {
		validator = &baseValidator{}
		subnetValidators[nodeID] = validator
	}
	return validator
}

func (v *baseValidators) pruneValidator(subnetID ids.ID, nodeID ids.NodeID) {
	subnetValidators, ok := v.validators[subnetID]
	if !ok {
		return
	}
	validator, ok := subnetValidators[nodeID]
	if !ok {
		return
	}
	if validator.staker != nil {
		return
	}
	if validator.delegators != nil && validator.delegators.Len() > 0 {
		return
	}
	delete(subnetValidators, nodeID)
	if len(subnetValidators) == 0 {
		delete(v.validators, subnetID)
	}
}

func (v *baseValidators) getOrCreateValidatorDiff(subnetID ids.ID, nodeID ids.NodeID) *diffValidator {
	subnetValidatorDiffs, ok := v.validatorDiffs[subnetID]
	if !ok {
		subnetValidatorDiffs = make(map[ids.NodeID]*diffValidator)
		v.validatorDiffs[subnetID] = subnetValidatorDiffs
	}
	validatorDiff, ok := subnetValidatorDiffs[nodeID]
	if !ok {
		validatorDiff = &diffValidator{}
		subnetValidatorDiffs[nodeID] = validatorDiff
	}
	return validatorDiff
}

type diffValidators struct {
	validatorDiffs map[ids.ID]map[ids.NodeID]*diffValidator
	addedStakers   *btree.BTree
	deletedStakers map[ids.ID]*Staker
}

func (v *diffValidators) getOrCreateDiff(subnetID ids.ID, nodeID ids.NodeID) *diffValidator {
	if v.validatorDiffs == nil {
		v.validatorDiffs = make(map[ids.ID]map[ids.NodeID]*diffValidator)
	}
	subnetValidatorDiffs, ok := v.validatorDiffs[subnetID]
	if !ok {
		subnetValidatorDiffs = make(map[ids.NodeID]*diffValidator)
		v.validatorDiffs[subnetID] = subnetValidatorDiffs
	}
	validatorDiff, ok := subnetValidatorDiffs[nodeID]
	if !ok {
		validatorDiff = &diffValidator{}
		subnetValidatorDiffs[nodeID] = validatorDiff
	}
	return validatorDiff
}

func (v *diffValidators) GetStaker(subnetID ids.ID, nodeID ids.NodeID) (*Staker, bool) {
	subnetValidatorDiffs, ok := v.validatorDiffs[subnetID]
	if !ok {
		return nil, false
	}

	validatorDiff, ok := subnetValidatorDiffs[nodeID]
	if !ok {
		return nil, false
	}

	if !validatorDiff.stakerModified {
		return nil, false
	}

	if validatorDiff.stakerDeleted {
		return nil, true
	}
	return validatorDiff.staker, true
}

func (v *diffValidators) PutStaker(staker *Staker) {
	validatorDiff := v.getOrCreateDiff(staker.SubnetID, staker.NodeID)
	validatorDiff.stakerModified = true
	validatorDiff.stakerDeleted = false
	validatorDiff.staker = staker

	if v.addedStakers == nil {
		v.addedStakers = btree.New(defaultTreeDegree)
	}
	v.addedStakers.ReplaceOrInsert(staker)
}

func (v *diffValidators) DeleteStaker(staker *Staker) {
	validatorDiff := v.getOrCreateDiff(staker.SubnetID, staker.NodeID)
	validatorDiff.stakerModified = true
	validatorDiff.stakerDeleted = true
	validatorDiff.staker = staker

	if v.deletedStakers == nil {
		v.deletedStakers = make(map[ids.ID]*Staker)
	}
	v.deletedStakers[staker.TxID] = staker
}

func (v *diffValidators) GetDelegatorIterator(parentIterator StakerIterator, subnetID ids.ID, nodeID ids.NodeID) StakerIterator {
	var (
		addedDelegatorIterator = EmptyIterator
		deletedDelegators      map[ids.ID]*Staker
	)
	if subnetValidatorDiffs, ok := v.validatorDiffs[subnetID]; ok {
		if validatorDiff, ok := subnetValidatorDiffs[nodeID]; ok {
			addedDelegatorIterator = NewTreeIterator(validatorDiff.addedDelegators)
			deletedDelegators = validatorDiff.deletedDelegators
		}
	}

	return NewMaskedIterator(
		NewMultiIterator(
			parentIterator,
			addedDelegatorIterator,
		),
		deletedDelegators,
	)
}

func (v *diffValidators) PutDelegator(staker *Staker) {
	validatorDiff := v.getOrCreateDiff(staker.SubnetID, staker.NodeID)
	if validatorDiff.addedDelegators == nil {
		validatorDiff.addedDelegators = btree.New(defaultTreeDegree)
	}
	validatorDiff.addedDelegators.ReplaceOrInsert(staker)

	if v.addedStakers == nil {
		v.addedStakers = btree.New(defaultTreeDegree)
	}
	v.addedStakers.ReplaceOrInsert(staker)
}

func (v *diffValidators) DeleteDelegator(staker *Staker) {
	validatorDiff := v.getOrCreateDiff(staker.SubnetID, staker.NodeID)
	if validatorDiff.deletedDelegators == nil {
		validatorDiff.deletedDelegators = make(map[ids.ID]*Staker)
	}
	validatorDiff.deletedDelegators[staker.TxID] = staker

	if v.deletedStakers == nil {
		v.deletedStakers = make(map[ids.ID]*Staker)
	}
	v.deletedStakers[staker.TxID] = staker
}

func (v *diffValidators) GetStakerIterator(parentIterator StakerIterator) StakerIterator {
	return NewMaskedIterator(
		NewMultiIterator(
			parentIterator,
			NewTreeIterator(v.addedStakers),
		),
		v.deletedStakers,
	)
}

type diffValidator struct {
	stakerModified bool
	stakerDeleted  bool
	staker         *Staker

	addedDelegators   *btree.BTree
	deletedDelegators map[ids.ID]*Staker
}
