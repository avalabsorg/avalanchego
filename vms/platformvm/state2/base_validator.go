// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/google/btree"

	"github.com/ava-labs/avalanchego/ids"
)

var (
	_ Validator  = &baseValidator{}
	_ Validators = &baseValidators{}
)

type baseValidator struct {
	currentStaker          *Staker
	pendingStaker          *Staker
	currentDelegatorWeight uint64
	currentDelegators      *btree.BTree
	pendingDelegators      *btree.BTree
}

func (v *baseValidator) CurrentStaker() *Staker {
	return v.currentStaker
}

func (v *baseValidator) PendingStaker() *Staker {
	return v.pendingStaker
}

func (v *baseValidator) CurrentDelegatorWeight() uint64 {
	return v.currentDelegatorWeight
}

func (v *baseValidator) NewCurrentDelegatorIterator() StakerIterator {
	return NewTreeIterator(v.currentDelegators)
}

func (v *baseValidator) NewPendingDelegatorIterator() StakerIterator {
	return NewTreeIterator(v.pendingDelegators)
}

type baseValidators struct {
	// Representation of DB state
	validators         map[ids.ID]map[ids.NodeID]*baseValidator
	nextRewardedStaker *Staker
	currentStakers     *btree.BTree
	pendingStakers     *btree.BTree

	// Representation of pending changes
	currentStakersToAdd    []*Staker
	currentStakersToRemove []*Staker
	pendingStakersToAdd    []*Staker
	pendingStakersToRemove []*Staker
}

func (v *baseValidators) GetValidator(subnetID ids.ID, nodeID ids.NodeID) Validator {
	subnetValidators, ok := v.validators[subnetID]
	if !ok {
		return &baseValidator{}
	}
	validator, ok := subnetValidators[nodeID]
	if !ok {
		return &baseValidator{}
	}
	return validator
}

func (v *baseValidators) GetNextRewardedStaker() *Staker {
	return v.nextRewardedStaker
}

func (v *baseValidators) NewCurrentStakerIterator() StakerIterator {
	return NewTreeIterator(v.currentStakers)
}

func (v *baseValidators) NewPendingStakerIterator() StakerIterator {
	return NewTreeIterator(v.pendingStakers)
}

func (v *baseValidators) Update(
	currentStakersToAdd []*Staker,
	currentStakersToRemove []*Staker,
	pendingStakersToAdd []*Staker,
	pendingStakersToRemove []*Staker,
) {
	v.currentStakersToAdd = currentStakersToAdd
	v.currentStakersToRemove = currentStakersToRemove
	v.pendingStakersToAdd = pendingStakersToAdd
	v.pendingStakersToRemove = pendingStakersToRemove
}
