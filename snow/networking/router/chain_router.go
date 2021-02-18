// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package router

import (
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/networking/timeout"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/timer"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

const (
	defaultCPUInterval = 15 * time.Second
)

var (
	_ Router = &ChainRouter{}
)

type request struct {
	// When this request was registered
	time.Time
	// The type of request that was made
	constants.MsgType
}

// ChainRouter routes incoming messages from the validator network
// to the consensus engines that the messages are intended for.
// Note that consensus engines are uniquely identified by the ID of the chain
// that they are working on.
type ChainRouter struct {
	clock            timer.Clock
	log              logging.Logger
	lock             sync.Mutex
	chains           map[ids.ID]*Handler
	timeoutManager   *timeout.Manager
	gossiper         *timer.Repeater
	intervalNotifier *timer.Repeater
	closeTimeout     time.Duration
	peers            ids.ShortSet
	criticalChains   ids.Set
	onFatal          func()
	// Unique-ified request ID --> Time and type of message made
	requests map[ids.ID]request
}

// Initialize the router.
//
// When this router receives an incoming message, it cancels the timeout in
// [timeouts] associated with the request that caused the incoming message, if
// applicable.
//
// This router also fires a gossip event every [gossipFrequency] to the engine,
// notifying the engine it should gossip it's accepted set.
func (cr *ChainRouter) Initialize(
	nodeID ids.ShortID,
	log logging.Logger,
	timeoutManager *timeout.Manager,
	gossipFrequency time.Duration,
	closeTimeout time.Duration,
	criticalChains ids.Set,
	onFatal func(),
) {
	cr.log = log
	cr.chains = make(map[ids.ID]*Handler)
	cr.timeoutManager = timeoutManager
	cr.gossiper = timer.NewRepeater(cr.Gossip, gossipFrequency)
	cr.intervalNotifier = timer.NewRepeater(cr.EndInterval, defaultCPUInterval)
	cr.closeTimeout = closeTimeout
	cr.criticalChains = criticalChains
	cr.onFatal = onFatal
	cr.requests = make(map[ids.ID]request)

	cr.peers.Add(nodeID)

	go log.RecoverAndPanic(cr.gossiper.Dispatch)
	go log.RecoverAndPanic(cr.intervalNotifier.Dispatch)
}

// RegisterRequests marks that we should expect to receive a reply from the given validator
// regarding the given chain and the reply should have the given requestID.
// The type of message we sent the validator was [msgType].
// Every registered request must be cleared either by receiving a valid reply
// and passing it to the appropriate chain or by a call to GetFailed, GetAncestorsFailed, etc.
// This method registers a timeout that calls such methods if we don't get a reply in time.
func (cr *ChainRouter) RegisterRequest(
	validatorID ids.ShortID,
	chainID ids.ID,
	requestID uint32,
	msgType constants.MsgType,
) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	// Add to the set of unfulfilled requests
	cr.requests[uniqueRequestID] = request{Time: cr.clock.Time(), MsgType: msgType}
	cr.lock.Unlock()
	// Register a timeout to fire if we don't get a reply in time.
	var timeoutHandler func() // Called upon timeout
	switch msgType {
	case constants.PullQueryMsg, constants.PushQueryMsg:
		timeoutHandler = func() { cr.QueryFailed(validatorID, chainID, requestID) }
	case constants.GetMsg:
		timeoutHandler = func() { cr.GetFailed(validatorID, chainID, requestID) }
	case constants.GetAncestorsMsg:
		timeoutHandler = func() { cr.GetAncestorsFailed(validatorID, chainID, requestID) }
	case constants.GetAcceptedMsg:
		timeoutHandler = func() { cr.GetAcceptedFailed(validatorID, chainID, requestID) }
	case constants.GetAcceptedFrontierMsg:
		timeoutHandler = func() { cr.GetAcceptedFrontierFailed(validatorID, chainID, requestID) }
	default:
		// This should never happen
		cr.log.Error("expected message type to be one of GetMsg, PullQueryMsg, PushQueryMsg, GetAcceptedFrontierMsg, GetAcceptedMsg but got %s", msgType)
		return
	}
	cr.timeoutManager.RegisterRequest(validatorID, chainID, uniqueRequestID, timeoutHandler)
}

// Shutdown shuts down this router
func (cr *ChainRouter) Shutdown() {
	cr.log.Info("shutting down chain router")
	cr.lock.Lock()
	prevChains := cr.chains
	cr.chains = map[ids.ID]*Handler{}
	cr.lock.Unlock()

	cr.gossiper.Stop()
	cr.intervalNotifier.Stop()

	for _, chain := range prevChains {
		chain.Shutdown()
	}

	ticker := time.NewTicker(cr.closeTimeout)
	timedout := false
	for _, chain := range prevChains {
		select {
		case <-chain.closed:
		case <-ticker.C:
			timedout = true
		}
	}
	if timedout {
		cr.log.Warn("timed out while shutting down the chains")
	}
	ticker.Stop()
}

// AddChain registers the specified chain so that incoming
// messages can be routed to it
func (cr *ChainRouter) AddChain(chain *Handler) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	chainID := chain.Context().ChainID
	cr.log.Debug("registering chain %s with chain router", chainID)
	chain.toClose = func() { cr.RemoveChain(chainID) }
	cr.chains[chainID] = chain

	for validatorID := range cr.peers {
		chain.Connected(validatorID)
	}
}

// RemoveChain removes the specified chain so that incoming
// messages can't be routed to it
func (cr *ChainRouter) RemoveChain(chainID ids.ID) {
	cr.lock.Lock()
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("can't remove unknown chain %s", chainID)
		cr.lock.Unlock()
		return
	}
	delete(cr.chains, chainID)
	cr.lock.Unlock()

	chain.Shutdown()

	ticker := time.NewTicker(cr.closeTimeout)
	select {
	case <-chain.closed:
	case <-ticker.C:
		chain.Context().Log.Warn("timed out while shutting down")
	}
	ticker.Stop()

	if cr.onFatal != nil && cr.criticalChains.Contains(chainID) {
		go cr.onFatal()
	}
}

// GetAcceptedFrontier routes an incoming GetAcceptedFrontier request from the
// validator with ID [validatorID]  to the consensus engine working on the
// chain with ID [chainID]
func (cr *ChainRouter) GetAcceptedFrontier(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("GetAcceptedFrontier(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the message to the chain
	// Note that we don't check the return value. It's OK to drop this.
	chain.GetAcceptedFrontier(validatorID, requestID, deadline)
}

// AcceptedFrontier routes an incoming AcceptedFrontier request from the
// validator with ID [validatorID]  to the consensus engine working on the
// chain with ID [chainID]
func (cr *ChainRouter) AcceptedFrontier(validatorID ids.ShortID, chainID ids.ID, requestID uint32, containerIDs []ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("AcceptedFrontier(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerIDs)
		return
	}

	uniqueRequestID := createRequestID(validatorID, chainID, requestID)

	// Mark that an outstanding request has been fulfilled
	request, exists := cr.requests[uniqueRequestID]
	if !exists || request.MsgType != constants.GetAcceptedFrontierMsg {
		// We didn't request this message or we got back a reply of wrong type. Ignore.
		return
	}
	delete(cr.requests, uniqueRequestID)

	// Calculate how long it took [validatorID] to reply
	latency := cr.clock.Time().Sub(request.Time)

	// Tell the timeout manager we got a response
	cr.timeoutManager.RegisterResponse(validatorID, chainID, uniqueRequestID, constants.GetAcceptedFrontierMsg, latency)

	// Pass the response to the chain
	if !chain.AcceptedFrontier(validatorID, requestID, containerIDs) {
		// We weren't able to pass the response to the chain
		chain.GetAcceptedFrontierFailed(validatorID, requestID)
	}
}

// GetAcceptedFrontierFailed routes an incoming GetAcceptedFrontierFailed
// request from the validator with ID [validatorID] to the consensus engine
// working on the chain with ID [chainID]
func (cr *ChainRouter) GetAcceptedFrontierFailed(validatorID ids.ShortID, chainID ids.ID, requestID uint32) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Remove the outstanding request
	delete(cr.requests, uniqueRequestID)

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		// Should only happen if node is shutting down
		cr.log.Debug("GetAcceptedFrontierFailed(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the response to the chain
	chain.GetAcceptedFrontierFailed(validatorID, requestID)
}

// GetAccepted routes an incoming GetAccepted request from the
// validator with ID [validatorID]  to the consensus engine working on the
// chain with ID [chainID]
func (cr *ChainRouter) GetAccepted(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time, containerIDs []ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("GetAccepted(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerIDs)
		return
	}

	// Pass the message to the chain.
	// Note that we don't check the return value. It's OK to drop this.
	chain.GetAccepted(validatorID, requestID, deadline, containerIDs)
}

// Accepted routes an incoming Accepted request from the validator with ID
// [validatorID] to the consensus engine working on the chain with ID
// [chainID]
func (cr *ChainRouter) Accepted(validatorID ids.ShortID, chainID ids.ID, requestID uint32, containerIDs []ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("Accepted(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerIDs)
		return
	}

	uniqueRequestID := createRequestID(validatorID, chainID, requestID)

	// Mark that an outstanding request has been fulfilled
	request, exists := cr.requests[uniqueRequestID]
	if !exists || request.MsgType != constants.GetAcceptedMsg {
		// We didn't request this message or we got back a reply of wrong type. Ignore.
		return
	}
	delete(cr.requests, uniqueRequestID)

	// Calculate how long it took [validatorID] to reply
	latency := cr.clock.Time().Sub(request.Time)

	// Tell the timeout manager we got a response
	cr.timeoutManager.RegisterResponse(validatorID, chainID, uniqueRequestID, constants.GetAcceptedMsg, latency)

	// Pass the response to the chain
	if !chain.Accepted(validatorID, requestID, containerIDs) {
		// We weren't able to pass the response to the chain
		chain.GetAcceptedFailed(validatorID, requestID)
	}
}

// GetAcceptedFailed routes an incoming GetAcceptedFailed request from the
// validator with ID [validatorID]  to the consensus engine working on the
// chain with ID [chainID]
func (cr *ChainRouter) GetAcceptedFailed(validatorID ids.ShortID, chainID ids.ID, requestID uint32) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Remove the outstanding request
	delete(cr.requests, uniqueRequestID)

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		// Should only happen when shutting down
		cr.log.Debug("GetAcceptedFailed(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the response to the chain
	chain.GetAcceptedFailed(validatorID, requestID)
}

// GetAncestors routes an incoming GetAncestors message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
// The maximum number of ancestors to respond with is defined in snow/engine/commong/bootstrapper.go
func (cr *ChainRouter) GetAncestors(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time, containerID ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("GetAncestors(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the message to the chain.
	// Note that we don't check the return value. It's OK to drop this.
	chain.GetAncestors(validatorID, requestID, deadline, containerID)
}

// MultiPut routes an incoming MultiPut message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) MultiPut(validatorID ids.ShortID, chainID ids.ID, requestID uint32, containers [][]byte) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("MultiPut(%s, %s, %d, %d) dropped due to unknown chain", validatorID, chainID, requestID, len(containers))
		return
	}

	uniqueRequestID := createRequestID(validatorID, chainID, requestID)

	// Mark that an outstanding request has been fulfilled
	request, exists := cr.requests[uniqueRequestID]
	if !exists || request.MsgType != constants.GetAncestorsMsg {
		// We didn't request this message or we got back a reply of wrong type. Ignore.
		return
	}
	delete(cr.requests, uniqueRequestID)

	// Calculate how long it took [validatorID] to reply
	latency := cr.clock.Time().Sub(request.Time)

	// Tell the timeout manager we got a response
	cr.timeoutManager.RegisterResponse(validatorID, chainID, uniqueRequestID, constants.GetAncestorsMsg, latency)

	// Pass the response to the chain
	if !chain.MultiPut(validatorID, requestID, containers) {
		// We weren't able to pass the response to the chain
		chain.GetAncestorsFailed(validatorID, requestID)
	}
}

// GetAncestorsFailed routes an incoming GetAncestorsFailed message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) GetAncestorsFailed(validatorID ids.ShortID, chainID ids.ID, requestID uint32) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Remove the outstanding request
	delete(cr.requests, uniqueRequestID)

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		// Should only happen if shutting down
		cr.log.Debug("GetAncestorsFailed(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the response to the chain
	chain.GetAncestorsFailed(validatorID, requestID)
}

// Get routes an incoming Get request from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) Get(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time, containerID ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("Get(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerID)
		return
	}

	// Pass the message to the chain.
	// Note that we don't check the return value. It's OK to drop this.
	chain.Get(validatorID, requestID, deadline, containerID)
}

// Put routes an incoming Put request from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) Put(validatorID ids.ShortID, chainID ids.ID, requestID uint32, containerID ids.ID, container []byte) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		if requestID == constants.GossipMsgRequestID {
			cr.log.Verbo("Gossiped Put(%s, %s, %d, %s) dropped due to unknown chain. Container:",
				validatorID, chainID, requestID, containerID, formatting.DumpBytes{Bytes: container},
			)
		} else {
			cr.log.Debug("Put(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerID)
			cr.log.Verbo("container:\n%s", formatting.DumpBytes{Bytes: container})
		}
		return
	}

	// If this is a gossip message, pass to the chain
	if requestID == constants.GossipMsgRequestID {
		// Note that we don't check the return value.
		// It's ok to drop this message.
		chain.Put(validatorID, requestID, containerID, container)
		return
	}

	uniqueRequestID := createRequestID(validatorID, chainID, requestID)

	// Mark that an outstanding request has been fulfilled
	request, exists := cr.requests[uniqueRequestID]
	if !exists || request.MsgType != constants.GetMsg {
		// We didn't request this message or we got back a reply of wrong type. Ignore.
		return
	}
	delete(cr.requests, uniqueRequestID)

	// Calculate how long it took [validatorID] to reply
	latency := cr.clock.Time().Sub(request.Time)

	// Tell the timeout manager we got a response
	cr.timeoutManager.RegisterResponse(validatorID, chainID, uniqueRequestID, constants.GetMsg, latency)

	// Pass the response to the chain
	if !chain.Put(validatorID, requestID, containerID, container) {
		// We weren't able to pass the response to the chain
		chain.GetFailed(validatorID, requestID)
	}
}

// GetFailed routes an incoming GetFailed message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) GetFailed(validatorID ids.ShortID, chainID ids.ID, requestID uint32) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Remove the outstanding request
	delete(cr.requests, uniqueRequestID)

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("GetFailed(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the response to the chain
	chain.GetFailed(validatorID, requestID)
}

// PushQuery routes an incoming PushQuery request from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) PushQuery(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time, containerID ids.ID, container []byte) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("PushQuery(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerID)
		cr.log.Verbo("container:\n%s", formatting.DumpBytes{Bytes: container})
		return
	}

	// Pass the message to the chain.
	// Note that we don't check the return value. It's OK to drop this.
	chain.PushQuery(validatorID, requestID, deadline, containerID, container)
}

// PullQuery routes an incoming PullQuery request from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) PullQuery(validatorID ids.ShortID, chainID ids.ID, requestID uint32, deadline time.Time, containerID ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("PullQuery(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, containerID)
		return
	}

	// Pass the message to the chain.
	// Note that we don't check the return value. It's OK to drop this.
	chain.PullQuery(validatorID, requestID, deadline, containerID)
}

// Chits routes an incoming Chits message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) Chits(validatorID ids.ShortID, chainID ids.ID, requestID uint32, votes []ids.ID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Get the chain, if it exists
	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("Chits(%s, %s, %d, %s) dropped due to unknown chain", validatorID, chainID, requestID, votes)
		return
	}

	uniqueRequestID := createRequestID(validatorID, chainID, requestID)

	// Mark that an outstanding request has been fulfilled
	request, exists := cr.requests[uniqueRequestID]
	if !exists || (request.MsgType != constants.PullQueryMsg && request.MsgType != constants.PushQueryMsg) {
		// We didn't request this message or we got back a reply of wrong type. Ignore.
		return
	}
	delete(cr.requests, uniqueRequestID)

	// Calculate how long it took [validatorID] to reply
	latency := cr.clock.Time().Sub(request.Time)

	// Tell the timeout manager we got a response
	cr.timeoutManager.RegisterResponse(validatorID, chainID, uniqueRequestID, request.MsgType, latency)

	// Pass the response to the chain
	if !chain.Chits(validatorID, requestID, votes) {
		// We weren't able to pass the response to the chain
		chain.QueryFailed(validatorID, requestID)
	}
}

// QueryFailed routes an incoming QueryFailed message from the validator with ID [validatorID]
// to the consensus engine working on the chain with ID [chainID]
func (cr *ChainRouter) QueryFailed(validatorID ids.ShortID, chainID ids.ID, requestID uint32) {
	uniqueRequestID := createRequestID(validatorID, chainID, requestID)
	cr.lock.Lock()
	defer cr.lock.Unlock()

	// Remove the outstanding request
	delete(cr.requests, uniqueRequestID)

	chain, exists := cr.chains[chainID]
	if !exists {
		cr.log.Debug("QueryFailed(%s, %s, %d) dropped due to unknown chain", validatorID, chainID, requestID)
		return
	}

	// Pass the response to the chain
	chain.QueryFailed(validatorID, requestID)
}

// Connected routes an incoming notification that a validator was just connected
func (cr *ChainRouter) Connected(validatorID ids.ShortID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	cr.peers.Add(validatorID)
	for _, chain := range cr.chains {
		chain.Connected(validatorID)
	}
}

// Disconnected routes an incoming notification that a validator was connected
func (cr *ChainRouter) Disconnected(validatorID ids.ShortID) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	cr.peers.Remove(validatorID)
	for _, chain := range cr.chains {
		chain.Disconnected(validatorID)
	}
}

// Gossip accepted containers
func (cr *ChainRouter) Gossip() {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	for _, chain := range cr.chains {
		chain.Gossip()
	}
}

// EndInterval notifies the chains that the current CPU interval has ended
func (cr *ChainRouter) EndInterval() {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	for _, chain := range cr.chains {
		chain.endInterval()
	}
}

func createRequestID(validatorID ids.ShortID, chainID ids.ID, requestID uint32) ids.ID {
	p := wrappers.Packer{Bytes: make([]byte, wrappers.IntLen)}
	p.PackInt(requestID)
	return hashing.ByteArraysToHash256Array(validatorID[:], chainID[:], p.Bytes)
}
