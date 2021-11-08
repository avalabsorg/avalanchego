// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package bootstrap

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/common/queue"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/metric"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

// Parameters for delaying bootstrapping to avoid potential CPU burns
const bootstrappingDelay = 10 * time.Second

var (
	_ SnowBootstrapper = &bootstrapper{}

	errUnexpectedTimeout = errors.New("unexpected timeout fired")
)

type Config struct {
	common.Config

	// Blocked tracks operations that are blocked on blocks
	Blocked *queue.JobsWithMissing

	VM block.ChainVM

	Bootstrapped func()
}

type SnowBootstrapper interface {
	common.Engine
	common.Bootstrapable

	Start(startReqID uint32) error
}

func New(
	config Config,
	onFinished func(lastReqID uint32) error,
	namespace string,
	registerer prometheus.Registerer,
) (SnowBootstrapper, error) {
	return newBootstrapper(
		config,
		onFinished,
		namespace,
		registerer,
	)
}

type bootstrapper struct {
	common.Bootstrapper
	common.Fetcher
	metrics
	getAncestorsBlks metric.Averager

	// Greatest height of the blocks passed in ForceAccepted
	tipHeight uint64
	// Height of the last accepted block when bootstrapping starts
	startingHeight uint64
	// Blocks passed into ForceAccepted
	startingAcceptedFrontier ids.Set

	// Blocked tracks operations that are blocked on blocks
	Blocked *queue.JobsWithMissing

	VM block.ChainVM

	Bootstrapped func()

	// number of state transitions executed
	executedStateTransitions int

	parser *parser

	awaitingTimeout bool
}

// new this engine.
func newBootstrapper(
	config Config,
	onFinished func(lastReqID uint32) error,
	namespace string,
	registerer prometheus.Registerer,
) (*bootstrapper, error) {
	res := &bootstrapper{}
	res.Blocked = config.Blocked
	res.VM = config.VM
	res.Bootstrapped = config.Bootstrapped
	res.OnFinished = onFinished
	res.executedStateTransitions = math.MaxInt32
	res.startingAcceptedFrontier = ids.Set{}

	lastAcceptedID, err := res.VM.LastAccepted()
	if err != nil {
		return nil, fmt.Errorf("couldn't get last accepted ID: %s", err)
	}
	lastAccepted, err := res.VM.GetBlock(lastAcceptedID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get last accepted block: %s", err)
	}
	res.startingHeight = lastAccepted.Height()

	if err := res.metrics.Initialize(namespace, registerer); err != nil {
		return nil, err
	}

	errs := wrappers.Errs{}
	res.getAncestorsBlks = metric.NewAveragerWithErrs(
		namespace,
		"get_ancestors_blks",
		"blocks fetched in a call to GetAncestors",
		registerer,
		&errs,
	)

	res.parser = &parser{
		log:         config.Ctx.Log,
		numAccepted: res.numAccepted,
		numDropped:  res.numDropped,
		vm:          res.VM,
	}
	if err := res.Blocked.SetParser(res.parser); err != nil {
		return nil, err
	}

	config.Bootstrapable = res
	if err := res.Bootstrapper.Initialize(config.Config); err != nil {
		return nil, err
	}

	return res, nil
}

func (b *bootstrapper) Start(startReqID uint32) error {
	b.Ctx.Log.Info("Starting bootstrap...")
	b.RequestID = startReqID

	if b.Config.StartupAlpha > 0 {
		return nil
	}

	return b.Bootstrapper.Startup()
}

// CurrentAcceptedFrontier returns the last accepted block
func (b *bootstrapper) CurrentAcceptedFrontier() ([]ids.ID, error) {
	lastAccepted, err := b.VM.LastAccepted()
	return []ids.ID{lastAccepted}, err
}

// FilterAccepted returns the blocks in [containerIDs] that we have accepted
func (b *bootstrapper) FilterAccepted(containerIDs []ids.ID) []ids.ID {
	acceptedIDs := make([]ids.ID, 0, len(containerIDs))
	for _, blkID := range containerIDs {
		if blk, err := b.VM.GetBlock(blkID); err == nil && blk.Status() == choices.Accepted {
			acceptedIDs = append(acceptedIDs, blkID)
		}
	}
	return acceptedIDs
}

func (b *bootstrapper) ForceAccepted(acceptedContainerIDs []ids.ID) error {
	if err := b.VM.Bootstrapping(); err != nil {
		return fmt.Errorf("failed to notify VM that bootstrapping has started: %w",
			err)
	}

	b.NumFetched = 0

	pendingContainerIDs := b.Blocked.MissingIDs()

	// Append the list of accepted container IDs to pendingContainerIDs to ensure
	// we iterate over every container that must be traversed.
	pendingContainerIDs = append(pendingContainerIDs, acceptedContainerIDs...)
	toProcess := make([]snowman.Block, 0, len(acceptedContainerIDs))
	b.Ctx.Log.Debug("Starting bootstrapping with %d pending blocks and %d from the accepted frontier",
		len(pendingContainerIDs), len(acceptedContainerIDs))
	for _, blkID := range pendingContainerIDs {
		b.startingAcceptedFrontier.Add(blkID)
		if blk, err := b.VM.GetBlock(blkID); err == nil {
			if height := blk.Height(); height > b.tipHeight {
				b.tipHeight = height
			}
			if blk.Status() == choices.Accepted {
				b.Blocked.RemoveMissingID(blkID)
			} else {
				toProcess = append(toProcess, blk)
			}
		} else {
			b.Blocked.AddMissingID(blkID)
			if err := b.fetch(blkID); err != nil {
				return err
			}
		}
	}

	// Process received blocks
	for _, blk := range toProcess {
		if err := b.process(blk, nil); err != nil {
			return err
		}
	}

	if numPending := b.Blocked.NumMissingIDs(); numPending == 0 {
		return b.checkFinish()
	}
	return nil
}

// Get block [blkID] and its ancestors from a validator
func (b *bootstrapper) fetch(blkID ids.ID) error {
	// Make sure we haven't already requested this block
	if b.OutstandingRequests.Contains(blkID) {
		return nil
	}

	// Make sure we don't already have this block
	if _, err := b.VM.GetBlock(blkID); err == nil {
		if numPending := b.Blocked.NumMissingIDs(); numPending == 0 {
			return b.checkFinish()
		}
		return nil
	}

	validators, err := b.Beacons.Sample(1) // validator to send request to
	if err != nil {
		return fmt.Errorf("dropping request for %s as there are no validators", blkID)
	}
	validatorID := validators[0].ID()
	b.RequestID++

	b.OutstandingRequests.Add(validatorID, b.RequestID, blkID)
	b.Sender.SendGetAncestors(validatorID, b.RequestID, blkID) // request block and ancestors
	return nil
}

// GetAncestors implements the Engine interface
func (b *bootstrapper) GetAncestors(vdr ids.ShortID, requestID uint32, blkID ids.ID) error {
	ancestorsBytes, err := block.GetAncestors(
		b.VM,
		blkID,
		b.MultiputMaxContainersSent,
		constants.MaxContainersLen,
		b.MaxTimeGetAncestors,
	)
	if err != nil {
		b.Ctx.Log.Verbo("couldn't get ancestors with %s. Dropping GetAncestors(%s, %d, %s)",
			err, vdr, requestID, blkID)
		return nil
	}

	b.getAncestorsBlks.Observe(float64(len(ancestorsBytes)))
	b.Sender.SendMultiPut(vdr, requestID, ancestorsBytes)
	return nil
}

// MultiPut handles the receipt of multiple containers. Should be received in response to a GetAncestors message to [vdr]
// with request ID [requestID]
func (b *bootstrapper) MultiPut(vdr ids.ShortID, requestID uint32, blks [][]byte) error {
	lenBlks := len(blks)
	if lenBlks == 0 {
		b.Ctx.Log.Debug("MultiPut(%s, %d) contains no blocks", vdr, requestID)
		return b.GetAncestorsFailed(vdr, requestID)
	}
	if lenBlks > b.MultiputMaxContainersReceived {
		blks = blks[:b.MultiputMaxContainersReceived]
		b.Ctx.Log.Debug("ignoring %d containers in multiput(%s, %d)", lenBlks-b.MultiputMaxContainersReceived, vdr, requestID)
	}

	// Make sure this is in response to a request we made
	wantedBlkID, ok := b.OutstandingRequests.Remove(vdr, requestID)
	if !ok { // this message isn't in response to a request we made
		b.Ctx.Log.Debug("received unexpected MultiPut from %s with ID %d", vdr, requestID)
		return nil
	}

	blocks, err := block.BatchedParseBlock(b.VM, blks)
	if err != nil { // the provided blocks couldn't be parsed
		b.Ctx.Log.Debug("failed to parse blocks in MultiPut from %s with ID %d", vdr, requestID)
		return b.fetch(wantedBlkID)
	}

	if len(blocks) == 0 {
		b.Ctx.Log.Debug("parsing blocks returned an empty set of blocks from %s with ID %d", vdr, requestID)
		return b.fetch(wantedBlkID)
	}

	requestedBlock := blocks[0]
	if actualID := requestedBlock.ID(); actualID != wantedBlkID {
		b.Ctx.Log.Debug("expected the first block to be the requested block, %s, but is %s",
			wantedBlkID, actualID)
		return b.fetch(wantedBlkID)
	}

	blockSet := make(map[ids.ID]snowman.Block, len(blocks))
	for _, block := range blocks[1:] {
		blockSet[block.ID()] = block
	}
	return b.process(requestedBlock, blockSet)
}

// GetAncestorsFailed is called when a GetAncestors message we sent fails
func (b *bootstrapper) GetAncestorsFailed(vdr ids.ShortID, requestID uint32) error {
	blkID, ok := b.OutstandingRequests.Remove(vdr, requestID)
	if !ok {
		b.Ctx.Log.Debug("GetAncestorsFailed(%s, %d) called but there was no outstanding request to this validator with this ID",
			vdr, requestID)
		return nil
	}
	// Send another request for this
	return b.fetch(blkID)
}

func (b *bootstrapper) Timeout() error {
	if !b.awaitingTimeout {
		return errUnexpectedTimeout
	}
	b.awaitingTimeout = false

	if !b.Subnet.IsBootstrapped() {
		return b.RestartBootstrap(true)
	}
	return b.finish()
}

// process a block
func (b *bootstrapper) process(blk snowman.Block, processingBlocks map[ids.ID]snowman.Block) error {
	status := blk.Status()
	blkID := blk.ID()
	blkHeight := blk.Height()
	if blkHeight > b.tipHeight && b.startingAcceptedFrontier.Contains(blkID) {
		b.tipHeight = blkHeight
	}

	for status == choices.Processing {
		if b.Halted() {
			return nil
		}

		b.Blocked.RemoveMissingID(blkID)

		pushed, err := b.Blocked.Push(&blockJob{
			parser:      b.parser,
			numAccepted: b.numAccepted,
			numDropped:  b.numDropped,
			blk:         blk,
			vm:          b.VM,
		})
		if err != nil {
			return err
		}

		// Traverse to the next block regardless of if the block is pushed
		blkID = blk.Parent()
		processingBlock, ok := processingBlocks[blkID]
		// first check processing blocks
		if ok {
			blk = processingBlock
			status = blk.Status()
		} else {
			// if not available in processing blocks, get block
			blk, err = b.VM.GetBlock(blkID)
			if err != nil {
				status = choices.Unknown
			} else {
				status = blk.Status()
			}
		}

		if !pushed {
			// If this block is already on the queue, then we can stop
			// traversing here.
			break
		}

		b.numFetched.Inc()
		b.NumFetched++                                      // Progress tracker
		if b.NumFetched%common.StatusUpdateFrequency == 0 { // Periodically print progress
			if !b.Restarted {
				b.Ctx.Log.Info("fetched %d of %d blocks", b.NumFetched, b.tipHeight-b.startingHeight)
			} else {
				b.Ctx.Log.Debug("fetched %d of %d blocks", b.NumFetched, b.tipHeight-b.startingHeight)
			}
		}
	}

	switch status {
	case choices.Unknown:
		b.Blocked.AddMissingID(blkID)
		if err := b.fetch(blkID); err != nil {
			return err
		}
	case choices.Rejected: // Should never happen
		return fmt.Errorf("bootstrapping wants to accept %s, however it was previously rejected", blkID)
	}

	if err := b.Blocked.Commit(); err != nil {
		return err
	}

	if numPending := b.Blocked.NumMissingIDs(); numPending == 0 {
		return b.checkFinish()
	}
	return nil
}

// checkFinish repeatedly executes pending transactions and requests new frontier vertices until there aren't any new ones
// after which it finishes the bootstrap process
func (b *bootstrapper) checkFinish() error {
	if b.IsBootstrapped() || b.awaitingTimeout {
		return nil
	}

	if !b.Restarted {
		b.Ctx.Log.Info("bootstrapping fetched %d blocks. Executing state transitions...", b.NumFetched)
	} else {
		b.Ctx.Log.Debug("bootstrapping fetched %d blocks. Executing state transitions...", b.NumFetched)
	}

	executedBlocks, err := b.Blocked.ExecuteAll(b.Ctx, b, b.Restarted, b.Ctx.ConsensusDispatcher, b.Ctx.DecisionDispatcher)
	if err != nil || b.Halted() {
		return err
	}

	previouslyExecuted := b.executedStateTransitions
	b.executedStateTransitions = executedBlocks

	// Note that executedBlocks < c*previouslyExecuted ( 0 <= c < 1 ) is enforced
	// so that the bootstrapping process will terminate even as new blocks are
	// being issued.
	if b.RetryBootstrap && executedBlocks > 0 && executedBlocks < previouslyExecuted/2 {
		return b.RestartBootstrap(true)
	}

	// If there is an additional callback, notify them that this chain has been
	// synced.
	if b.Bootstrapped != nil {
		b.Bootstrapped()
	}

	// Notify the subnet that this chain is synced
	b.Subnet.Bootstrapped(b.Ctx.ChainID)

	// If the subnet hasn't finished bootstrapping, this chain should remain
	// syncing.
	if !b.Subnet.IsBootstrapped() {
		if !b.Restarted {
			b.Ctx.Log.Info("waiting for the remaining chains in this subnet to finish syncing")
		} else {
			b.Ctx.Log.Debug("waiting for the remaining chains in this subnet to finish syncing")
		}
		// Restart bootstrapping after [bootstrappingDelay] to keep up to date
		// on the latest tip.
		b.Timer.RegisterTimeout(bootstrappingDelay)
		b.awaitingTimeout = true
		return nil
	}

	return b.finish()
}

func (b *bootstrapper) finish() error {
	if err := b.VM.Bootstrapped(); err != nil {
		return fmt.Errorf("failed to notify VM that bootstrapping has finished: %w",
			err)
	}

	// Start consensus
	if err := b.OnFinished(b.RequestID); err != nil {
		return err
	}
	b.Ctx.SetState(snow.NormalOp)
	return nil
}

// Connected implements the Engine interface.
func (b *bootstrapper) Connected(nodeID ids.ShortID) error {
	if err := b.VM.Connected(nodeID); err != nil {
		return err
	}

	return b.Bootstrapper.Connected(nodeID)
}

// Disconnected implements the Engine interface.
func (b *bootstrapper) Disconnected(nodeID ids.ShortID) error {
	if err := b.VM.Disconnected(nodeID); err != nil {
		return err
	}

	return b.Bootstrapper.Disconnected(nodeID)
}

// AppHandler interface
func (b *bootstrapper) AppRequest(nodeID ids.ShortID, requestID uint32, deadline time.Time, request []byte) error {
	b.Ctx.Log.Debug("dropping AppRequest(%s, %d) due to bootstrapping", nodeID, requestID)
	return nil
}

func (b *bootstrapper) AppResponse(nodeID ids.ShortID, requestID uint32, response []byte) error {
	b.Ctx.Log.Debug("dropping AppResponse(%s, %d) due to bootstrapping", nodeID, requestID)
	return nil
}

func (b *bootstrapper) AppRequestFailed(nodeID ids.ShortID, requestID uint32) error {
	b.Ctx.Log.Debug("dropping AppRequestFailed(%s, %d) due to bootstrapping", nodeID, requestID)
	return nil
}

func (b *bootstrapper) AppGossip(nodeID ids.ShortID, msg []byte) error {
	b.Ctx.Log.Debug("dropping AppGossip(%s) due to bootstrapping", nodeID)
	return nil
}

// End of AppHandler interface

func (b *bootstrapper) Get(validatorID ids.ShortID, requestID uint32, containerID ids.ID) error {
	b.Ctx.Log.Debug("Received Get message from (%s) during bootstrap. Dropping it", validatorID)
	return nil
}

func (b *bootstrapper) Put(vdr ids.ShortID, requestID uint32, blkID ids.ID, blkBytes []byte) error {
	// bootstrapping isn't done --> we didn't send any gets --> this put is invalid
	if requestID == constants.GossipMsgRequestID {
		b.Ctx.Log.Verbo("dropping gossip Put(%s, %d, %s) due to bootstrapping", vdr, requestID, blkID)
	} else {
		b.Ctx.Log.Debug("dropping Put(%s, %d, %s) due to bootstrapping", vdr, requestID, blkID)
	}
	return nil
}

func (b *bootstrapper) GetFailed(validatorID ids.ShortID, requestID uint32) error {
	// not done bootstrapping --> didn't send a get --> this message is invalid
	b.Ctx.Log.Debug("dropping GetFailed(%s, %d) due to bootstrapping", validatorID, requestID)
	return nil
}

func (b *bootstrapper) GetVM() common.VM {
	return b.VM
}

func (b *bootstrapper) Gossip() error {
	b.Ctx.Log.Debug("No Gossip during bootstrap. Dropping it")
	return nil
}

func (b *bootstrapper) Notify(common.Message) error {
	b.Ctx.Log.Debug("dropping Notify due to bootstrapping")
	return nil
}

func (b *bootstrapper) Shutdown() error {
	b.Ctx.Log.Debug("Called Shutdown during bootstrap. Doing nothing for now")
	return nil
}

func (b *bootstrapper) HealthCheck() (interface{}, error) {
	vmIntf, vmErr := b.VM.HealthCheck()
	intf := map[string]interface{}{
		"consensus": struct{}{},
		"vm":        vmIntf,
	}
	return intf, vmErr
}

// QueryHandler interface
func (b *bootstrapper) PullQuery(vdr ids.ShortID, requestID uint32, blkID ids.ID) error {
	b.Ctx.Log.Debug("dropping PullQuery(%s, %d, %s) due to bootstrapping", vdr, requestID, blkID)
	return nil
}

func (b *bootstrapper) PushQuery(vdr ids.ShortID, requestID uint32, blkID ids.ID, blkBytes []byte) error {
	b.Ctx.Log.Debug("dropping PushQuery(%s, %d, %s) due to bootstrapping", vdr, requestID, blkID)
	return nil
}

func (b *bootstrapper) Chits(vdr ids.ShortID, requestID uint32, votes []ids.ID) error {
	b.Ctx.Log.Debug("dropping Chits(%s, %d) due to bootstrapping", vdr, requestID)
	return nil
}

func (b *bootstrapper) QueryFailed(vdr ids.ShortID, requestID uint32) error {
	b.Ctx.Log.Debug("dropping QueryFailed(%s, %d) due to bootstrapping", vdr, requestID)
	return nil
}

// End of QueryHandler interface
