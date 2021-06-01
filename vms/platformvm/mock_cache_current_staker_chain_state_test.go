// Code generated by mockery v1.0.0. DO NOT EDIT.

package platformvm

import (
	ids "github.com/ava-labs/avalanchego/ids"
	mock "github.com/stretchr/testify/mock"

	validators "github.com/ava-labs/avalanchego/snow/validators"
)

// currentStakerChainState is an autogenerated mock type for the currentStakerChainState type
type mockCurrentStakerChainState struct {
	mock.Mock
}

// Apply provides a mock function with given fields: _a0
func (_m *mockCurrentStakerChainState) Apply(_a0 InternalState) {
	_m.Called(_a0)
}

// DeleteNextStaker provides a mock function with given fields:
func (_m *mockCurrentStakerChainState) DeleteNextStaker() (currentStakerChainState, error) {
	ret := _m.Called()

	var r0 currentStakerChainState
	if rf, ok := ret.Get(0).(func() currentStakerChainState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(currentStakerChainState)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNextStaker provides a mock function with given fields:
func (_m *mockCurrentStakerChainState) GetNextStaker() (*Tx, uint64, error) {
	ret := _m.Called()

	var r0 *Tx
	if rf, ok := ret.Get(0).(func() *Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Tx)
		}
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func() uint64); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(uint64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetStaker provides a mock function with given fields: txID
func (_m *mockCurrentStakerChainState) GetStaker(txID ids.ID) (*Tx, uint64, error) {
	ret := _m.Called(txID)

	var r0 *Tx
	if rf, ok := ret.Get(0).(func(ids.ID) *Tx); ok {
		r0 = rf(txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Tx)
		}
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func(ids.ID) uint64); ok {
		r1 = rf(txID)
	} else {
		r1 = ret.Get(1).(uint64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(ids.ID) error); ok {
		r2 = rf(txID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetValidator provides a mock function with given fields: nodeID
func (_m *mockCurrentStakerChainState) GetValidator(nodeID ids.ShortID) (currentValidator, error) {
	ret := _m.Called(nodeID)

	var r0 currentValidator
	if rf, ok := ret.Get(0).(func(ids.ShortID) currentValidator); ok {
		r0 = rf(nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(currentValidator)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ShortID) error); ok {
		r1 = rf(nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Stakers provides a mock function with given fields:
func (_m *mockCurrentStakerChainState) Stakers() []*Tx {
	ret := _m.Called()

	var r0 []*Tx
	if rf, ok := ret.Get(0).(func() []*Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Tx)
		}
	}

	return r0
}

// UpdateStakers provides a mock function with given fields: addValidators, addDelegators, addSubnetValidators, numTxsToRemove
func (_m *mockCurrentStakerChainState) UpdateStakers(addValidators []*validatorReward, addDelegators []*validatorReward, addSubnetValidators []*Tx, numTxsToRemove int) (currentStakerChainState, error) {
	ret := _m.Called(addValidators, addDelegators, addSubnetValidators, numTxsToRemove)

	var r0 currentStakerChainState
	if rf, ok := ret.Get(0).(func([]*validatorReward, []*validatorReward, []*Tx, int) currentStakerChainState); ok {
		r0 = rf(addValidators, addDelegators, addSubnetValidators, numTxsToRemove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(currentStakerChainState)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*validatorReward, []*validatorReward, []*Tx, int) error); ok {
		r1 = rf(addValidators, addDelegators, addSubnetValidators, numTxsToRemove)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidatorSet provides a mock function with given fields: subnetID
func (_m *mockCurrentStakerChainState) ValidatorSet(subnetID ids.ID) (validators.Set, error) {
	ret := _m.Called(subnetID)

	var r0 validators.Set
	if rf, ok := ret.Get(0).(func(ids.ID) validators.Set); ok {
		r0 = rf(subnetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(validators.Set)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ID) error); ok {
		r1 = rf(subnetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
