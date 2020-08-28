// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validators

import (
	"sync"

	"github.com/ava-labs/gecko/ids"
)

// Manager holds the validator set of each subnet
type Manager interface {
	// Set a subnet's validator set
	Set(ids.ID, Set)

	// AddWeight adds weight to a given validator on the given subnet
	AddWeight(ids.ID, ids.ShortID, uint64) error

	// RemoveWeight removes weight from a given validator on a given subnet
	RemoveWeight(ids.ID, ids.ShortID, uint64)

	// GetValidators returns the validator set for the given subnet
	// Returns false if the subnet doesn't exist
	GetValidators(ids.ID) (Set, bool)
}

// NewManager returns a new, empty manager
func NewManager() Manager {
	return &manager{
		subnetToVdrs: make(map[[32]byte]Set),
	}
}

// manager implements Manager
type manager struct {
	lock sync.Mutex
	// Key: Subnet ID
	// Value: The validators that validate the subnet
	subnetToVdrs map[[32]byte]Set
}

func (m *manager) Set(subnetID ids.ID, vdrSet Set) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.subnetToVdrs[subnetID.Key()] = vdrSet
}

// AddWeight implements the Manager interface.
func (m *manager) AddWeight(subnetID ids.ID, vdrID ids.ShortID, weight uint64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	subnetIDKey := subnetID.Key()

	vdrs, ok := m.subnetToVdrs[subnetIDKey]
	if !ok {
		vdrs = NewBestSet(5)
		if err := vdrs.AddWeight(vdrID, weight); err != nil {
			return err
		}
	} else {
		vdrs.AddWeight(vdrID, weight)
	}
	m.subnetToVdrs[subnetIDKey] = vdrs
	return nil
}

// RemoveValidatorSet implements the Manager interface.
func (m *manager) RemoveWeight(subnetID ids.ID, vdrID ids.ShortID, weight uint64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	vdrs, ok := m.subnetToVdrs[subnetID.Key()]
	if ok {
		vdrs.RemoveWeight(vdrID, weight)
	}
}

// GetValidatorSet implements the Manager interface.
func (m *manager) GetValidators(subnetID ids.ID) (Set, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	vdrs, ok := m.subnetToVdrs[subnetID.Key()]
	return vdrs, ok
}
