// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/consensus/avalanche"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/networking/sender"
)

var _ Subnet = (*subnet)(nil)

// Subnet keeps track of the currently bootstrapping chains in a subnet. If no
// chains in the subnet are currently bootstrapping, the subnet is considered
// bootstrapped.
type Subnet interface {
	common.Subnet

	afterBootstrapped() chan struct{}

	addChain(chainID ids.ID)
	removeChain(chainID ids.ID)
}

type SubnetConfig struct {
	sender.GossipConfig

	// ValidatorOnly indicates that this Subnet's Chains are available to only subnet validators.
	ValidatorOnly       bool                 `json:"validatorOnly" yaml:"validatorOnly"`
	ConsensusParameters avalanche.Parameters `json:"consensusParameters" yaml:"consensusParameters"`
}

type subnet struct {
	lock             sync.RWMutex
	bootstrapping    ids.Set
	once             sync.Once
	bootstrappedSema chan struct{}
}

func newSubnet() Subnet {
	return &subnet{
		bootstrappedSema: make(chan struct{}),
	}
}

func (s *subnet) IsBootstrapped() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.bootstrapping.Len() == 0
}

func (s *subnet) Bootstrapped(chainID ids.ID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.bootstrapping.Remove(chainID)
	if s.bootstrapping.Len() > 0 {
		return
	}

	s.once.Do(func() {
		close(s.bootstrappedSema)
	})
}

func (s *subnet) afterBootstrapped() chan struct{} {
	return s.bootstrappedSema
}

func (s *subnet) addChain(chainID ids.ID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.bootstrapping.Add(chainID)
}

func (s *subnet) removeChain(chainID ids.ID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.bootstrapping.Remove(chainID)
}
