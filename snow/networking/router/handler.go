// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package router

import (
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/snow"

	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/networking/tracker"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	"github.com/ava-labs/avalanchego/utils/uptime"
)

var errDuplicatedContainerID = errors.New("inbound message contains duplicated container ID")

// Handler passes incoming messages from the network to the consensus engine.
// (Actually, it receives the incoming messages from a ChainRouter, but same difference.)
type Handler struct {
	ctx *snow.Context
	// Useful for faking time in tests
	clock   mockable.Clock
	mc      message.Creator
	metrics handlerMetrics
	// The validator set that validates this chain
	validators validators.Set

	fastSyncer   common.Engine
	bootstrapper common.Engine
	engine       common.Engine

	// Closed when this handler and [engine] are done shutting down
	closed chan struct{}
	// Receives messages from the VM
	msgFromVMChan <-chan common.Message
	// Tracks CPU time spent processing messages from each node
	cpuTracker tracker.TimeTracker
	// Called in a goroutine when this handler/engine shuts down.
	// May be nil.
	onCloseF            func()
	unprocessedMsgsCond *sync.Cond
	// Holds messages that [engine] hasn't processed yet.
	// [unprocessedMsgsCond.L] must be held while accessing [unprocessedMsgs].
	unprocessedMsgs unprocessedMsgs
	closing         utils.AtomicBool
}

// Initialize this consensus handler
// [engine] must be initialized before initializing this handler
func NewHandler(
	mc message.Creator,
	ctx *snow.Context,
	validators validators.Set,
	msgFromVMChan <-chan common.Message,
	metricsNamespace string,
	metricsRegisterer prometheus.Registerer,
) (*Handler, error) {
	h := &Handler{
		ctx:                 ctx,
		mc:                  mc,
		closed:              make(chan struct{}),
		msgFromVMChan:       msgFromVMChan,
		validators:          validators,
		unprocessedMsgsCond: sync.NewCond(&sync.Mutex{}),
		cpuTracker:          tracker.NewCPUTracker(uptime.IntervalFactory{}, defaultCPUInterval),
	}

	if err := h.metrics.Initialize(metricsNamespace, metricsRegisterer); err != nil {
		return nil, fmt.Errorf("initializing handler metrics errored with: %s", err)
	}
	var err error
	h.unprocessedMsgs, err = newUnprocessedMsgs(h.ctx.Log, h.validators, h.cpuTracker, metricsNamespace, metricsRegisterer)
	return h, err
}

func (h *Handler) RegisterFastSyncer(fastSyncer common.Engine) {
	h.fastSyncer = fastSyncer
}

func (h *Handler) RegisterBootstrap(smbootstrap common.Engine) {
	h.bootstrapper = smbootstrap
}

func (h *Handler) RegisterEngine(engine common.Engine) {
	h.engine = engine
}

func (h *Handler) OnDoneFastSyncing(lastReqID uint32) error {
	lastReqID++
	return h.bootstrapper.Start(lastReqID)
}

func (h *Handler) OnDoneBootstrapping(lastReqID uint32) error {
	lastReqID++
	return h.engine.Start(lastReqID)
}

// Context of this Handler
func (h *Handler) Context() *snow.Context { return h.ctx }

// Engine returns the engine this handler dispatches to
func (h *Handler) Engine() common.Engine { return h.engine }

// SetEngine sets the engine for this handler to dispatch to
func (h *Handler) SetEngine(engine common.Engine) { h.engine = engine }

// Push the message onto the handler's queue
func (h *Handler) Push(msg message.InboundMessage) {
	nodeID := msg.NodeID()
	if nodeID == ids.ShortEmpty {
		// This should never happen
		h.ctx.Log.Warn("message does not have node ID of sender. Message: %s", msg)
	}

	h.unprocessedMsgsCond.L.Lock()
	defer h.unprocessedMsgsCond.L.Unlock()

	h.unprocessedMsgs.Push(msg)
	h.unprocessedMsgsCond.Signal()
}

// Dispatch waits for incoming messages from the router
// and, when they arrive, sends them to the consensus engine
func (h *Handler) Dispatch() {
	defer h.shutdown()

	// Handle messages from the VM
	go h.dispatchInternal()

	// Handle messages from the router
	for {
		// Wait until there is an unprocessed message
		h.unprocessedMsgsCond.L.Lock()
		for {
			if closing := h.closing.GetValue(); closing {
				h.unprocessedMsgsCond.L.Unlock()
				return
			}
			if h.unprocessedMsgs.Len() == 0 {
				// Signalled in [h.push] and [h.StartShutdown]
				h.unprocessedMsgsCond.Wait()
				continue
			}
			break
		}

		// Get the next message we should process
		msg := h.unprocessedMsgs.Pop()
		h.unprocessedMsgsCond.L.Unlock()

		// If this message's deadline has passed, don't process it.
		if expirationTime := msg.ExpirationTime(); !expirationTime.IsZero() && h.clock.Time().After(expirationTime) {
			nodeID := msg.NodeID()
			h.ctx.Log.Verbo("Dropping message from %s%s due to timeout. msg: %s",
				constants.NodeIDPrefix, nodeID, msg)
			h.metrics.expired.Inc()
			msg.OnFinishedHandling()
			continue
		}

		// Process the message.
		// If there was an error, shut down this chain
		if err := h.handleMsg(msg); err != nil {
			h.ctx.Log.Fatal("chain shutting down due to error %q while processing message: %s",
				err, msg)
			h.StartShutdown()
			return
		}
	}
}

// IsPeriodic returns true if this message is of a type that is sent on a
// periodic basis.
func isPeriodic(inMsg message.InboundMessage) bool {
	op := inMsg.Op()
	if op == message.AppGossip || op == message.GossipRequest {
		return true
	}
	if op != message.Put {
		return false
	}

	reqID := inMsg.Get(message.RequestID).(uint32)
	return reqID == constants.GossipMsgRequestID
}

// Dispatch a message to the consensus engine.
func (h *Handler) handleMsg(msg message.InboundMessage) error {
	startTime := h.clock.Time()

	isPeriodic := isPeriodic(msg)
	if isPeriodic {
		h.ctx.Log.Verbo("Forwarding message to consensus: %s", msg)
	} else {
		h.ctx.Log.Debug("Forwarding message to consensus: %s", msg)
	}

	h.ctx.Lock.Lock()
	defer h.ctx.Lock.Unlock()

	var (
		err error
		op  = msg.Op()
	)
	switch op {
	case message.Notify:
		vmMsg := msg.Get(message.VMMessage).(uint32)
		if h.ctx.IsBootstrapped() {
			err = h.engine.Notify(common.Message(vmMsg))
		} else {
			err = h.bootstrapper.Notify(common.Message(vmMsg))
		}

	case message.GossipRequest:
		err = h.engine.Gossip()

	case message.Timeout:
		err = h.bootstrapper.Timeout()

	default:
		err = h.handleConsensusMsg(msg)
	}

	endTime := h.clock.Time()
	// If the message was caused by another node, track their CPU time.
	if op != message.Notify && op != message.GossipRequest && op != message.Timeout {
		nodeID := msg.NodeID()
		h.cpuTracker.UtilizeTime(nodeID, startTime, endTime)
	}

	// Track how long the operation took.
	histogram := h.metrics.messages[op]
	// TODO: should not be needed
	if histogram == nil {
		h.ctx.Log.Warn("could not find metric map for message type %s", op.String())
	} else {
		histogram.Observe(float64(endTime.Sub(startTime)))
	}

	msg.OnFinishedHandling()

	if isPeriodic {
		h.ctx.Log.Verbo("Finished handling message: %s", op)
	} else {
		h.ctx.Log.Debug("Finished handling message: %s", op)
	}
	return err
}

// Assumes [h.ctx.Lock] is locked
// Relevant fields in msgs must be validated before being dispatched to the engine.
// An invalid msg is logged and dropped silently since err would cause a chain shutdown.
func (h *Handler) handleConsensusMsg(msg message.InboundMessage) error {
	nodeID := msg.NodeID()
	switch msg.Op() {
	case message.GetAcceptedFrontier:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.bootstrapper.GetAcceptedFrontier(nodeID, reqID)

	case message.AcceptedFrontier:
		reqID := msg.Get(message.RequestID).(uint32)
		containerIDs, err := getContainerIDs(msg)
		if err != nil {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: %s",
				msg.Op(), nodeID, h.ctx.ChainID, reqID, err)
			return nil
		}

		return h.bootstrapper.AcceptedFrontier(nodeID, reqID, containerIDs)

	case message.GetAcceptedFrontierFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.bootstrapper.GetAcceptedFrontierFailed(nodeID, reqID)

	case message.GetAccepted:
		reqID := msg.Get(message.RequestID).(uint32)
		containerIDs, err := getContainerIDs(msg)
		if err != nil {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: %s",
				msg.Op(), nodeID, h.ctx.ChainID, reqID, err)
			return nil
		}
		return h.bootstrapper.GetAccepted(nodeID, reqID, containerIDs)

	case message.Accepted:
		reqID := msg.Get(message.RequestID).(uint32)
		containerIDs, err := getContainerIDs(msg)
		if err != nil {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: %s",
				msg.Op(), nodeID, h.ctx.ChainID, reqID, err)
			return nil
		}

		return h.bootstrapper.Accepted(nodeID, reqID, containerIDs)

	case message.GetAcceptedFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.bootstrapper.GetAcceptedFailed(nodeID, reqID)

	case message.GetAncestors:
		reqID := msg.Get(message.RequestID).(uint32)
		containerID, err := ids.ToID(msg.Get(message.ContainerID).([]byte))
		if err != nil {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: %s",
				msg.Op(), nodeID, h.ctx.ChainID, reqID, err)
			return nil
		}

		return h.bootstrapper.GetAncestors(nodeID, reqID, containerID)

	case message.GetAncestorsFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.bootstrapper.GetAncestorsFailed(nodeID, reqID)

	case message.MultiPut:
		reqID := msg.Get(message.RequestID).(uint32)
		containers := msg.Get(message.MultiContainerBytes).([][]byte)
		return h.bootstrapper.MultiPut(nodeID, reqID, containers)

	case message.Get:
		reqID := msg.Get(message.RequestID).(uint32)
		containerID, err := ids.ToID(msg.Get(message.ContainerID).([]byte))
		h.ctx.Log.AssertNoError(err)
		return h.engine.Get(nodeID, reqID, containerID)

	case message.GetFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		if h.ctx.IsBootstrapped() {
			return h.engine.GetFailed(nodeID, reqID)
		}
		return h.bootstrapper.GetFailed(nodeID, reqID)

	case message.Put:
		reqID := msg.Get(message.RequestID).(uint32)
		containerID, err := ids.ToID(msg.Get(message.ContainerID).([]byte))
		h.ctx.Log.AssertNoError(err)
		container, ok := msg.Get(message.ContainerBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse ContainerBytes",
				msg.Op(), nodeID, h.ctx.ChainID, reqID)
			return nil
		}

		if h.ctx.IsBootstrapped() {
			return h.engine.Put(nodeID, reqID, containerID, container)
		}
		return h.bootstrapper.Put(nodeID, reqID, containerID, container)

	case message.PushQuery:
		reqID := msg.Get(message.RequestID).(uint32)
		containerID, err := ids.ToID(msg.Get(message.ContainerID).([]byte))
		h.ctx.Log.AssertNoError(err)
		container, ok := msg.Get(message.ContainerBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse ContainerBytes",
				msg.Op(), nodeID, h.ctx.ChainID, reqID)
			return nil
		}

		if h.ctx.IsBootstrapped() {
			return h.engine.PushQuery(nodeID, reqID, containerID, container)
		}
		return h.bootstrapper.PushQuery(nodeID, reqID, containerID, container)

	case message.PullQuery:
		reqID := msg.Get(message.RequestID).(uint32)
		containerID, err := ids.ToID(msg.Get(message.ContainerID).([]byte))
		h.ctx.Log.AssertNoError(err)

		if h.ctx.IsBootstrapped() {
			return h.engine.PullQuery(nodeID, reqID, containerID)
		}
		return h.bootstrapper.PullQuery(nodeID, reqID, containerID)

	case message.Chits:
		reqID := msg.Get(message.RequestID).(uint32)
		votes, err := getContainerIDs(msg)
		if err != nil {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: %s",
				msg.Op(), nodeID, h.ctx.ChainID, reqID, err)
			return nil
		}

		if h.ctx.IsBootstrapped() {
			return h.engine.Chits(nodeID, reqID, votes)
		}
		return h.bootstrapper.Chits(nodeID, reqID, votes)

	case message.QueryFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		if h.ctx.IsBootstrapped() {
			return h.engine.QueryFailed(nodeID, reqID)
		}
		return h.bootstrapper.QueryFailed(nodeID, reqID)

	case message.Connected:
		return h.bootstrapper.Connected(nodeID)

	case message.Disconnected:
		return h.bootstrapper.Disconnected(nodeID)

	case message.AppRequest:
		reqID := msg.Get(message.RequestID).(uint32)
		appBytes, ok := msg.Get(message.AppBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse AppBytes",
				msg.Op(), nodeID, h.ctx.ChainID, reqID)
			return nil
		}
		if h.ctx.IsBootstrapped() {
			return h.engine.AppRequest(nodeID, reqID, msg.ExpirationTime(), appBytes)
		}
		return h.bootstrapper.AppRequest(nodeID, reqID, msg.ExpirationTime(), appBytes)

	case message.AppResponse:
		reqID := msg.Get(message.RequestID).(uint32)
		appBytes, ok := msg.Get(message.AppBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse AppBytes",
				msg.Op(), nodeID, h.ctx.ChainID, reqID)
			return nil
		}
		if h.ctx.IsBootstrapped() {
			return h.engine.AppResponse(nodeID, reqID, appBytes)
		}
		return h.bootstrapper.AppResponse(nodeID, reqID, appBytes)

	case message.AppRequestFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		if h.ctx.IsBootstrapped() {
			return h.engine.AppRequestFailed(nodeID, reqID)
		}
		return h.bootstrapper.AppRequestFailed(nodeID, reqID)

	case message.AppGossip:
		appBytes, ok := msg.Get(message.AppBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse AppBytes",
				msg.Op(), nodeID, h.ctx.ChainID, constants.GossipMsgRequestID)
			return nil
		}

		if h.ctx.IsBootstrapped() {
			return h.engine.AppGossip(nodeID, appBytes)
		}
		return h.bootstrapper.AppGossip(nodeID, appBytes)

	case message.GetStateSummaryFrontier:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.fastSyncer.GetStateSummaryFrontier(nodeID, reqID)

	case message.StateSummaryFrontier:
		reqID := msg.Get(message.RequestID).(uint32)
		summary, ok := msg.Get(message.ContainerBytes).([]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse ContainerBytes",
				msg.Op(), nodeID, h.engine.Context().ChainID, reqID)
			return nil
		}
		return h.fastSyncer.StateSummaryFrontier(nodeID, reqID, summary)

	case message.GetStateSummaryFrontierFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.fastSyncer.GetStateSummaryFrontierFailed(nodeID, reqID)

	case message.GetAcceptedStateSummary:
		reqID := msg.Get(message.RequestID).(uint32)
		summaries, ok := msg.Get(message.MultiContainerBytes).([][]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse MultiContainerBytes",
				msg.Op(), nodeID, h.engine.Context().ChainID, reqID)
			return nil
		}
		return h.fastSyncer.GetAcceptedStateSummary(nodeID, reqID, summaries)

	case message.AcceptedStateSummary:
		reqID := msg.Get(message.RequestID).(uint32)
		summaries, ok := msg.Get(message.MultiContainerBytes).([][]byte)
		if !ok {
			h.ctx.Log.Debug("Malformed message %s from (%s, %s, %d) dropped. Error: could not parse MultiContainerBytes",
				msg.Op(), nodeID, h.engine.Context().ChainID, reqID)
			return nil
		}
		return h.fastSyncer.AcceptedStateSummary(nodeID, reqID, summaries)

	case message.GetAcceptedStateSummaryFailed:
		reqID := msg.Get(message.RequestID).(uint32)
		return h.fastSyncer.GetAcceptedStateSummaryFailed(nodeID, reqID)

	default:
		h.ctx.Log.Warn("Attempt to submit to engine unhandled consensus msg %s from from (%s, %s). Dropping it",
			msg.Op(), nodeID, h.ctx.ChainID)
		return nil
	}
}

// Timeout passes a new timeout notification to the consensus engine
func (h *Handler) Timeout() {
	msg := h.mc.InternalTimeout(h.ctx.NodeID)
	h.Push(msg)
}

// Gossip passes a gossip request to the consensus engine
func (h *Handler) Gossip() {
	if !h.ctx.IsBootstrapped() {
		// Shouldn't send gossiping messages while the chain is bootstrapping
		return
	}

	inMsg := h.mc.InternalGossipRequest(h.ctx.NodeID)
	h.Push(inMsg)
}

// StartShutdown starts the shutdown process for this handler/engine.
// [h] must never be invoked again after calling this method.
// This method causes [shutdown] to eventually be called.
// [h.closed] is closed when this handler/engine are done shutting down.
func (h *Handler) StartShutdown() {
	// Must hold [h.unprocessedMsgsCond.L] here to ensure
	// there's no race condition in Dispatch where we check
	// the value of [h.closing].
	h.unprocessedMsgsCond.L.Lock()
	h.closing.SetValue(true)
	h.unprocessedMsgsCond.L.Unlock()

	// If we're waiting in [Dispatch] wake up.
	h.unprocessedMsgsCond.Signal()
	// Don't process any more bootstrap messages.
	// If [h.engine] is processing a bootstrap message, stop.
	// We do this because if we didn't, and the engine was in the
	// middle of executing state transitions during bootstrapping,
	// we wouldn't be able to grab [h.ctx.Lock] until the engine
	// finished executing state transitions, which may take a long time.
	// As a result, the router would time out on shutting down this chain.
	h.bootstrapper.Halt()
}

// Calls [h.engine.Shutdown] and [h.onCloseF]; closes [h.closed].
func (h *Handler) shutdown() {
	h.ctx.Lock.Lock()
	defer h.ctx.Lock.Unlock()

	startTime := h.clock.Time()
	if err := h.engine.Shutdown(); err != nil {
		h.ctx.Log.Error("Error while shutting down the chain: %s", err)
	}
	if h.onCloseF != nil {
		go h.onCloseF()
	}
	endTime := h.clock.Time()
	h.metrics.shutdown.Observe(float64(endTime.Sub(startTime)))
	close(h.closed)
}

func (h *Handler) dispatchInternal() {
	for {
		select {
		case <-h.closed:
			return
		case msg := <-h.msgFromVMChan:
			if closing := h.closing.GetValue(); closing {
				return
			}
			// handle a message from the VM
			inMsg := h.mc.InternalVMMessage(h.ctx.NodeID, uint32(msg))
			h.Push(inMsg)
		}
	}
}

func (h *Handler) endInterval() {
	now := h.clock.Time()
	h.cpuTracker.EndInterval(now)
}

// if subnet is validator only and this is not a validator or self, returns false.
func (h *Handler) isValidator(nodeID ids.ShortID) bool {
	return !h.ctx.IsValidatorOnly() || nodeID == h.ctx.NodeID || h.validators.Contains(nodeID)
}

func getContainerIDs(msg message.InboundMessage) ([]ids.ID, error) {
	containerIDsBytes := msg.Get(message.ContainerIDs).([][]byte)
	res := make([]ids.ID, len(containerIDsBytes))
	idSet := ids.NewSet(len(containerIDsBytes))
	for i, containerIDBytes := range containerIDsBytes {
		containerID, err := ids.ToID(containerIDBytes)
		if err != nil {
			return nil, err
		}
		if idSet.Contains(containerID) {
			return nil, errDuplicatedContainerID
		}
		res[i] = containerID
		idSet.Add(containerID)
	}
	return res, nil
}
