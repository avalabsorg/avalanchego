// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sender

import (
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
	"github.com/ava-labs/gecko/snow/networking/router"
	"github.com/ava-labs/gecko/snow/networking/timeout"
)

// Sender sends consensus messages to other validators
type Sender struct {
	ctx      *snow.Context
	sender   ExternalSender // Actually does the sending over the network
	router   router.Router
	timeouts *timeout.Manager
}

// Initialize this sender
func (s *Sender) Initialize(ctx *snow.Context, sender ExternalSender, router router.Router, timeouts *timeout.Manager) {
	s.ctx = ctx
	s.sender = sender
	s.router = router
	s.timeouts = timeouts
}

// Context of this sender
func (s *Sender) Context() *snow.Context { return s.ctx }

// GetAcceptedFrontier ...
func (s *Sender) GetAcceptedFrontier(validatorIDs ids.ShortSet, requestID uint32) {
	if validatorIDs.Contains(s.ctx.NodeID) {
		validatorIDs.Remove(s.ctx.NodeID)
		go s.router.GetAcceptedFrontier(s.ctx.NodeID, s.ctx.ChainID, requestID)
	}
	validatorList := validatorIDs.List()
	for _, validatorID := range validatorList {
		vID := validatorID
		s.timeouts.Register(validatorID, s.ctx.ChainID, requestID, func() {
			s.router.GetAcceptedFrontierFailed(vID, s.ctx.ChainID, requestID)
		})
	}
	s.sender.GetAcceptedFrontier(validatorIDs, s.ctx.ChainID, requestID)
}

// AcceptedFrontier ...
func (s *Sender) AcceptedFrontier(validatorID ids.ShortID, requestID uint32, containerIDs ids.Set) {
	if validatorID.Equals(s.ctx.NodeID) {
		go s.router.AcceptedFrontier(validatorID, s.ctx.ChainID, requestID, containerIDs)
		return
	}
	s.sender.AcceptedFrontier(validatorID, s.ctx.ChainID, requestID, containerIDs)
}

// GetAccepted ...
func (s *Sender) GetAccepted(validatorIDs ids.ShortSet, requestID uint32, containerIDs ids.Set) {
	if validatorIDs.Contains(s.ctx.NodeID) {
		validatorIDs.Remove(s.ctx.NodeID)
		go s.router.GetAccepted(s.ctx.NodeID, s.ctx.ChainID, requestID, containerIDs)
	}
	validatorList := validatorIDs.List()
	for _, validatorID := range validatorList {
		vID := validatorID
		s.timeouts.Register(validatorID, s.ctx.ChainID, requestID, func() {
			s.router.GetAcceptedFailed(vID, s.ctx.ChainID, requestID)
		})
	}
	s.sender.GetAccepted(validatorIDs, s.ctx.ChainID, requestID, containerIDs)
}

// Accepted ...
func (s *Sender) Accepted(validatorID ids.ShortID, requestID uint32, containerIDs ids.Set) {
	if validatorID.Equals(s.ctx.NodeID) {
		go s.router.Accepted(validatorID, s.ctx.ChainID, requestID, containerIDs)
		return
	}
	s.sender.Accepted(validatorID, s.ctx.ChainID, requestID, containerIDs)
}

// Get sends a Get message to the consensus engine running on the specified
// chain to the specified validator. The Get message signifies that this
// consensus engine would like the recipient to send this consensus engine the
// specified container.
func (s *Sender) Get(validatorID ids.ShortID, requestID uint32, containerID ids.ID) {
	s.ctx.Log.Verbo("Sending Get to validator %s. RequestID: %d. ContainerID: %s", validatorID, requestID, containerID)
	// Add a timeout -- if we don't get a response before the timeout expires,
	// send this consensus engine a GetFailed message
	s.timeouts.Register(validatorID, s.ctx.ChainID, requestID, func() {
		s.router.GetFailed(validatorID, s.ctx.ChainID, requestID)
	})
	s.sender.Get(validatorID, s.ctx.ChainID, requestID, containerID)
}

// Put sends a Put message to the consensus engine running on the specified chain
// on the specified validator.
// The Put message signifies that this consensus engine is giving to the recipient
// the contents of the specified container.
func (s *Sender) Put(validatorID ids.ShortID, requestID uint32, containerID ids.ID, container []byte) {
	s.ctx.Log.Verbo("Sending Put to validator %s. RequestID: %d. ContainerID: %s", validatorID, requestID, containerID)
	s.sender.Put(validatorID, s.ctx.ChainID, requestID, containerID, container)
}

// PushQuery sends a PushQuery message to the consensus engines running on the specified chains
// on the specified validators.
// The PushQuery message signifies that this consensus engine would like each validator to send
// their preferred frontier given the existence of the specified container.
func (s *Sender) PushQuery(validatorIDs ids.ShortSet, requestID uint32, containerID ids.ID, container []byte) {
	s.ctx.Log.Verbo("Sending PushQuery to validators %v. RequestID: %d. ContainerID: %s", validatorIDs, requestID, containerID)
	// If one of the validators in [validatorIDs] is myself, send this message directly
	// to my own router rather than sending it over the network
	if validatorIDs.Contains(s.ctx.NodeID) { // One of the validators in [validatorIDs] was myself
		validatorIDs.Remove(s.ctx.NodeID)
		// We use a goroutine to avoid a deadlock in the case where the consensus engine queries itself.
		// The flow of execution in that case is handler --> engine --> sender --> chain router --> handler
		// If this were not a goroutine, then we would deadlock here when [handler].msgs is full
		go s.router.PushQuery(s.ctx.NodeID, s.ctx.ChainID, requestID, containerID, container)
	}
	validatorList := validatorIDs.List() // Convert set to list for easier iteration
	for _, validatorID := range validatorList {
		vID := validatorID
		s.timeouts.Register(validatorID, s.ctx.ChainID, requestID, func() {
			s.router.QueryFailed(vID, s.ctx.ChainID, requestID)
		})
	}
	s.sender.PushQuery(validatorIDs, s.ctx.ChainID, requestID, containerID, container)
}

// PullQuery sends a PullQuery message to the consensus engines running on the specified chains
// on the specified validators.
// The PullQuery message signifies that this consensus engine would like each validator to send
// their preferred frontier.
func (s *Sender) PullQuery(validatorIDs ids.ShortSet, requestID uint32, containerID ids.ID) {
	s.ctx.Log.Verbo("Sending PullQuery. RequestID: %d. ContainerID: %s", requestID, containerID)
	// If one of the validators in [validatorIDs] is myself, send this message directly
	// to my own router rather than sending it over the network
	if validatorIDs.Contains(s.ctx.NodeID) { // One of the validators in [validatorIDs] was myself
		validatorIDs.Remove(s.ctx.NodeID)
		// We use a goroutine to avoid a deadlock in the case where the consensus engine queries itself.
		// The flow of execution in that case is handler --> engine --> sender --> chain router --> handler
		// If this were not a goroutine, then we would deadlock when [handler].msgs is full
		go s.router.PullQuery(s.ctx.NodeID, s.ctx.ChainID, requestID, containerID)
	}
	validatorList := validatorIDs.List() // Convert set to list for easier iteration
	for _, validatorID := range validatorList {
		vID := validatorID
		s.timeouts.Register(validatorID, s.ctx.ChainID, requestID, func() {
			s.router.QueryFailed(vID, s.ctx.ChainID, requestID)
		})
	}
	s.sender.PullQuery(validatorIDs, s.ctx.ChainID, requestID, containerID)
}

// Chits sends chits
func (s *Sender) Chits(validatorID ids.ShortID, requestID uint32, votes ids.Set) {
	s.ctx.Log.Verbo("Sending Chits to validator %s. RequestID: %d. Votes: %s", validatorID, requestID, votes)
	// If [validatorID] is myself, send this message directly
	// to my own router rather than sending it over the network
	if validatorID.Equals(s.ctx.NodeID) {
		go s.router.Chits(validatorID, s.ctx.ChainID, requestID, votes)
		return
	}
	s.sender.Chits(validatorID, s.ctx.ChainID, requestID, votes)
}

// Gossip the provided container
func (s *Sender) Gossip(containerID ids.ID, container []byte) {
	s.ctx.Log.Verbo("Gossiping %s", containerID)
	s.sender.Gossip(s.ctx.ChainID, containerID, container)
}
