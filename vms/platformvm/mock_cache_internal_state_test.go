// Code generated by mockery v1.0.0. DO NOT EDIT.

package platformvm

import (
	database "github.com/ava-labs/avalanchego/database"
	avax "github.com/ava-labs/avalanchego/vms/components/avax"

	ids "github.com/ava-labs/avalanchego/ids"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

var _ InternalState = &MockInternalState{}

// InternalState is an autogenerated mock type for the InternalState type
type MockInternalState struct {
	mock.Mock
}

// Abort provides a mock function with given fields:
func (_m *MockInternalState) Abort() {
	_m.Called()
}

// AddBlock provides a mock function with given fields: block
func (_m *MockInternalState) AddBlock(block Block) {
	_m.Called(block)
}

// AddChain provides a mock function with given fields: createChainTx
func (_m *MockInternalState) AddChain(createChainTx *Tx) {
	_m.Called(createChainTx)
}

// AddCurrentStaker provides a mock function with given fields: tx, potentialReward
func (_m *MockInternalState) AddCurrentStaker(tx *Tx, potentialReward uint64) {
	_m.Called(tx, potentialReward)
}

// AddPendingStaker provides a mock function with given fields: tx
func (_m *MockInternalState) AddPendingStaker(tx *Tx) {
	_m.Called(tx)
}

// AddRewardUTXO provides a mock function with given fields: txID, utxo
func (_m *MockInternalState) AddRewardUTXO(txID ids.ID, utxo *avax.UTXO) {
	_m.Called(txID, utxo)
}

// AddSubnet provides a mock function with given fields: createSubnetTx
func (_m *MockInternalState) AddSubnet(createSubnetTx *Tx) {
	_m.Called(createSubnetTx)
}

// AddTx provides a mock function with given fields: tx, status
func (_m *MockInternalState) AddTx(tx *Tx, status Status) {
	_m.Called(tx, status)
}

// AddUTXO provides a mock function with given fields: utxo
func (_m *MockInternalState) AddUTXO(utxo *avax.UTXO) {
	_m.Called(utxo)
}

// Close provides a mock function with given fields:
func (_m *MockInternalState) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Commit provides a mock function with given fields:
func (_m *MockInternalState) Commit() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CommitBatch provides a mock function with given fields:
func (_m *MockInternalState) CommitBatch() (database.Batch, error) {
	ret := _m.Called()

	var r0 database.Batch
	if rf, ok := ret.Get(0).(func() database.Batch); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(database.Batch)
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

// CurrentStakerChainState provides a mock function with given fields:
func (_m *MockInternalState) CurrentStakerChainState() currentStakerChainState {
	ret := _m.Called()

	var r0 currentStakerChainState
	if rf, ok := ret.Get(0).(func() currentStakerChainState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(currentStakerChainState)
		}
	}

	return r0
}

// DeleteCurrentStaker provides a mock function with given fields: tx
func (_m *MockInternalState) DeleteCurrentStaker(tx *Tx) {
	_m.Called(tx)
}

// DeletePendingStaker provides a mock function with given fields: tx
func (_m *MockInternalState) DeletePendingStaker(tx *Tx) {
	_m.Called(tx)
}

// DeleteUTXO provides a mock function with given fields: utxoID
func (_m *MockInternalState) DeleteUTXO(utxoID ids.ID) {
	_m.Called(utxoID)
}

// GetBlock provides a mock function with given fields: blockID
func (_m *MockInternalState) GetBlock(blockID ids.ID) (Block, error) {
	ret := _m.Called(blockID)

	var r0 Block
	if rf, ok := ret.Get(0).(func(ids.ID) Block); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ID) error); ok {
		r1 = rf(blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetChains provides a mock function with given fields: subnetID
func (_m *MockInternalState) GetChains(subnetID ids.ID) ([]*Tx, error) {
	ret := _m.Called(subnetID)

	var r0 []*Tx
	if rf, ok := ret.Get(0).(func(ids.ID) []*Tx); ok {
		r0 = rf(subnetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Tx)
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

// GetCurrentSupply provides a mock function with given fields:
func (_m *MockInternalState) GetCurrentSupply() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetLastAccepted provides a mock function with given fields:
func (_m *MockInternalState) GetLastAccepted() ids.ID {
	ret := _m.Called()

	var r0 ids.ID
	if rf, ok := ret.Get(0).(func() ids.ID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ids.ID)
		}
	}

	return r0
}

// GetRewardUTXOs provides a mock function with given fields: txID
func (_m *MockInternalState) GetRewardUTXOs(txID ids.ID) ([]*avax.UTXO, error) {
	ret := _m.Called(txID)

	var r0 []*avax.UTXO
	if rf, ok := ret.Get(0).(func(ids.ID) []*avax.UTXO); ok {
		r0 = rf(txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*avax.UTXO)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ID) error); ok {
		r1 = rf(txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSubnets provides a mock function with given fields:
func (_m *MockInternalState) GetSubnets() ([]*Tx, error) {
	ret := _m.Called()

	var r0 []*Tx
	if rf, ok := ret.Get(0).(func() []*Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Tx)
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

// GetTimestamp provides a mock function with given fields:
func (_m *MockInternalState) GetTimestamp() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// GetTx provides a mock function with given fields: txID
func (_m *MockInternalState) GetTx(txID ids.ID) (*Tx, Status, error) {
	ret := _m.Called(txID)

	var r0 *Tx
	if rf, ok := ret.Get(0).(func(ids.ID) *Tx); ok {
		r0 = rf(txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Tx)
		}
	}

	var r1 Status
	if rf, ok := ret.Get(1).(func(ids.ID) Status); ok {
		r1 = rf(txID)
	} else {
		r1 = ret.Get(1).(Status)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(ids.ID) error); ok {
		r2 = rf(txID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetUTXO provides a mock function with given fields: utxoID
func (_m *MockInternalState) GetUTXO(utxoID ids.ID) (*avax.UTXO, error) {
	ret := _m.Called(utxoID)

	var r0 *avax.UTXO
	if rf, ok := ret.Get(0).(func(ids.ID) *avax.UTXO); ok {
		r0 = rf(utxoID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*avax.UTXO)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ID) error); ok {
		r1 = rf(utxoID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUptime provides a mock function with given fields: nodeID
func (_m *MockInternalState) GetUptime(nodeID ids.ShortID) (time.Duration, time.Time, error) {
	ret := _m.Called(nodeID)

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(ids.ShortID) time.Duration); ok {
		r0 = rf(nodeID)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	var r1 time.Time
	if rf, ok := ret.Get(1).(func(ids.ShortID) time.Time); ok {
		r1 = rf(nodeID)
	} else {
		r1 = ret.Get(1).(time.Time)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(ids.ShortID) error); ok {
		r2 = rf(nodeID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// PendingStakerChainState provides a mock function with given fields:
func (_m *MockInternalState) PendingStakerChainState() pendingStakerChainState {
	ret := _m.Called()

	var r0 pendingStakerChainState
	if rf, ok := ret.Get(0).(func() pendingStakerChainState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pendingStakerChainState)
		}
	}

	return r0
}

// SetCurrentStakerChainState provides a mock function with given fields: _a0
func (_m *MockInternalState) SetCurrentStakerChainState(_a0 currentStakerChainState) {
	_m.Called(_a0)
}

// SetCurrentSupply provides a mock function with given fields: _a0
func (_m *MockInternalState) SetCurrentSupply(_a0 uint64) {
	_m.Called(_a0)
}

// SetLastAccepted provides a mock function with given fields: _a0
func (_m *MockInternalState) SetLastAccepted(_a0 ids.ID) {
	_m.Called(_a0)
}

// SetMigrated provides a mock function with given fields:
func (_m *MockInternalState) SetMigrated() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPendingStakerChainState provides a mock function with given fields: _a0
func (_m *MockInternalState) SetPendingStakerChainState(_a0 pendingStakerChainState) {
	_m.Called(_a0)
}

// SetTimestamp provides a mock function with given fields: _a0
func (_m *MockInternalState) SetTimestamp(_a0 time.Time) {
	_m.Called(_a0)
}

// SetUptime provides a mock function with given fields: nodeID, upDuration, lastUpdated
func (_m *MockInternalState) SetUptime(nodeID ids.ShortID, upDuration time.Duration, lastUpdated time.Time) error {
	ret := _m.Called(nodeID, upDuration, lastUpdated)

	var r0 error
	if rf, ok := ret.Get(0).(func(ids.ShortID, time.Duration, time.Time) error); ok {
		r0 = rf(nodeID, upDuration, lastUpdated)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UTXOIDs provides a mock function with given fields: addr, start, limit
func (_m *MockInternalState) UTXOIDs(addr []byte, start ids.ID, limit int) ([]ids.ID, error) {
	ret := _m.Called(addr, start, limit)

	var r0 []ids.ID
	if rf, ok := ret.Get(0).(func([]byte, ids.ID, int) []ids.ID); ok {
		r0 = rf(addr, start, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]ids.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, ids.ID, int) error); ok {
		r1 = rf(addr, start, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
