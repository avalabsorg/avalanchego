// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/rpc/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/snow/uptime"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	"github.com/ava-labs/avalanchego/utils/window"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/version"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/api"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateful"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/fx"
	"github.com/ava-labs/avalanchego/vms/platformvm/reward"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/builder"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/executor"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/mempool"
	"github.com/ava-labs/avalanchego/vms/platformvm/utxo"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

var (
	_ block.ChainVM    = &VM{}
	_ secp256k1fx.VM   = &VM{}
	_ validators.State = &VM{}

	errWrongCacheType = errors.New("unexpectedly cached type")
)

const (
	validatorSetsCacheSize        = 64
	maxRecentlyAcceptedWindowSize = 256
	recentlyAcceptedWindowTTL     = 5 * time.Minute
)

type VM struct {
	Factory
	blockBuilder

	metrics
	avax.AddressManager
	avax.AtomicUTXOManager
	*network

	// Used to get time. Useful for faking time during tests.
	clock mockable.Clock

	uptimeManager uptime.Manager

	rewards reward.Calculator

	// The context of this vm
	ctx       *snow.Context
	dbManager manager.Manager

	internalState state.State
	utxoHandler   utxo.Handler

	// ID of the preferred block
	preferred ids.ID

	fx            fx.Fx
	codecRegistry codec.Registry

	// Bootstrapped remembers if this chain has finished bootstrapping or not
	bootstrapped utils.AtomicBool

	// Maps caches for each subnet that is currently whitelisted.
	// Key: Subnet ID
	// Value: cache mapping height -> validator set map
	validatorSetCaches map[ids.ID]cache.Cacher

	// Key: block ID
	// Value: the block
	currentBlocks map[ids.ID]stateful.Block

	// sliding window of blocks that were recently accepted
	recentlyAccepted *window.Window

	txBuilder         builder.TxBuilder
	txExecutorBackend executor.Backend
	blkVerifier       stateful.Verifier
}

// Initialize this blockchain.
// [vm.ChainManager] and [vm.vdrMgr] must be set before this function is called.
func (vm *VM) Initialize(
	ctx *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	upgradeBytes []byte,
	configBytes []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) error {
	var err error
	ctx.Log.Verbo("initializing platform chain")

	registerer := prometheus.NewRegistry()
	vm.ctx = ctx
	if err := ctx.Metrics.Register(registerer); err != nil {
		return err
	}

	vm.dbManager = dbManager

	vm.codecRegistry = linearcodec.NewDefault()
	vm.fx = &secp256k1fx.Fx{}
	if err := vm.fx.Initialize(vm); err != nil {
		return err
	}

	// Initialize metrics as soon as possible
	if err := vm.metrics.Initialize("", registerer, vm.WhitelistedSubnets); err != nil {
		return err
	}

	vm.validatorSetCaches = make(map[ids.ID]cache.Cacher)
	vm.recentlyAccepted = window.New(
		window.Config{
			Clock:   &vm.clock,
			MaxSize: maxRecentlyAcceptedWindowSize,
			TTL:     recentlyAcceptedWindowTTL,
		},
	)

	vm.rewards = reward.NewCalculator(vm.RewardConfig)
	vm.currentBlocks = make(map[ids.ID]stateful.Block)
	if vm.internalState, err = state.New(
		vm.dbManager.Current().Database,
		registerer,
		&vm.Config,
		vm.ctx,
		vm.metrics.localStake,
		vm.metrics.totalStake,
		vm.rewards,
		genesisBytes,
	); err != nil {
		return err
	}

	vm.AddressManager = avax.NewAddressManager(ctx)
	vm.AtomicUTXOManager = avax.NewAtomicUTXOManager(ctx.SharedMemory, txs.Codec)
	vm.utxoHandler = utxo.NewHandler(vm.ctx, &vm.clock, vm.internalState, vm.fx)
	vm.uptimeManager = uptime.NewManager(vm.internalState)
	vm.UptimeLockedCalculator.SetCalculator(&vm.bootstrapped, &ctx.Lock, vm.uptimeManager)

	vm.txBuilder = builder.NewTxBuilder(
		vm.ctx,
		vm.Config,
		&vm.clock,
		vm.fx,
		vm.internalState,
		vm.AtomicUTXOManager,
		vm.utxoHandler,
	)

	vm.txExecutorBackend = executor.Backend{
		Cfg:          &vm.Config,
		Ctx:          vm.ctx,
		Clk:          &vm.clock,
		Fx:           vm.fx,
		SpendHandler: vm.utxoHandler,
		UptimeMan:    vm.uptimeManager,
		Rewards:      vm.rewards,
		Bootstrapped: &vm.bootstrapped,
	}

	// Note: there is a circular dependency among mempool and blkBuilder
	// which is broken by mean of vm
	mempool, err := mempool.NewMempool("mempool", registerer, vm)
	if err != nil {
		return fmt.Errorf("failed to create mempool: %w", err)
	}
	vm.blkVerifier = NewBlockVerifier(
		mempool,
		vm.internalState,
		vm.txExecutorBackend,
		&vm.metrics,
		vm.recentlyAccepted,
	)

	if err := vm.blockBuilder.Initialize(
		mempool,
		vm,
		toEngine,
		registerer,
	); err != nil {
		return fmt.Errorf(
			"failed to initialize the block builder: %w",
			err,
		)
	}
	vm.network = newNetwork(vm.ApricotPhase4Time, appSender, vm)

	if err := vm.updateValidators(); err != nil {
		return fmt.Errorf(
			"failed to initialize validator sets: %w",
			err,
		)
	}

	// Create all of the chains that the database says exist
	if err := vm.initBlockchains(); err != nil {
		return fmt.Errorf(
			"failed to initialize blockchains: %w",
			err,
		)
	}

	// Build off the most recently accepted block
	lastAcceptedID := vm.internalState.GetLastAccepted()
	ctx.Log.Info("initializing last accepted block as %s", lastAcceptedID)
	return vm.SetPreference(lastAcceptedID)
}

// Create all chains that exist that this node validates.
func (vm *VM) initBlockchains() error {
	if err := vm.createSubnet(constants.PrimaryNetworkID); err != nil {
		return err
	}

	if vm.StakingEnabled {
		for subnetID := range vm.WhitelistedSubnets {
			if err := vm.createSubnet(subnetID); err != nil {
				return err
			}
		}
	} else {
		subnets, err := vm.internalState.GetSubnets()
		if err != nil {
			return err
		}
		for _, subnet := range subnets {
			if err := vm.createSubnet(subnet.ID()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Create the subnet with ID [subnetID]
func (vm *VM) createSubnet(subnetID ids.ID) error {
	chains, err := vm.internalState.GetChains(subnetID)
	if err != nil {
		return err
	}
	for _, chain := range chains {
		tx, ok := chain.Unsigned.(*txs.CreateChainTx)
		if !ok {
			return fmt.Errorf("expected tx type *txs.CreateChainTx but got %T", chain.Unsigned)
		}
		vm.Config.CreateChain(chain.ID(), tx)
	}
	return nil
}

// onBootstrapStarted marks this VM as bootstrapping
func (vm *VM) onBootstrapStarted() error {
	vm.bootstrapped.SetValue(false)
	return vm.fx.Bootstrapping()
}

// onNormalOperationsStarted marks this VM as bootstrapped
func (vm *VM) onNormalOperationsStarted() error {
	if vm.bootstrapped.GetValue() {
		return nil
	}
	vm.bootstrapped.SetValue(true)

	if err := vm.fx.Bootstrapped(); err != nil {
		return err
	}

	primaryValidatorSet, exist := vm.Validators.GetValidators(constants.PrimaryNetworkID)
	if !exist {
		return errNoPrimaryValidators
	}
	primaryValidators := primaryValidatorSet.List()

	validatorIDs := make([]ids.NodeID, len(primaryValidators))
	for i, vdr := range primaryValidators {
		validatorIDs[i] = vdr.ID()
	}

	if err := vm.uptimeManager.StartTracking(validatorIDs); err != nil {
		return err
	}
	return vm.internalState.Commit()
}

func (vm *VM) SetState(state snow.State) error {
	switch state {
	case snow.Bootstrapping:
		return vm.onBootstrapStarted()
	case snow.NormalOp:
		return vm.onNormalOperationsStarted()
	default:
		return snow.ErrUnknownState
	}
}

// Shutdown this blockchain
func (vm *VM) Shutdown() error {
	if vm.dbManager == nil {
		return nil
	}

	vm.blockBuilder.Shutdown()

	if vm.bootstrapped.GetValue() {
		primaryValidatorSet, exist := vm.Validators.GetValidators(constants.PrimaryNetworkID)
		if !exist {
			return errNoPrimaryValidators
		}
		primaryValidators := primaryValidatorSet.List()

		validatorIDs := make([]ids.NodeID, len(primaryValidators))
		for i, vdr := range primaryValidators {
			validatorIDs[i] = vdr.ID()
		}

		if err := vm.uptimeManager.Shutdown(validatorIDs); err != nil {
			return err
		}
		if err := vm.internalState.Commit(); err != nil {
			return err
		}
	}

	errs := wrappers.Errs{}
	errs.Add(
		vm.internalState.Close(),
		vm.dbManager.Close(),
	)
	return errs.Err
}

// BuildBlock builds a block to be added to consensus
func (vm *VM) BuildBlock() (snowman.Block, error) { return vm.blockBuilder.BuildBlock() }

func (vm *VM) ParseBlock(b []byte) (snowman.Block, error) {
	// Note: blocks to be parsed are not verified, so we must used stateless.Codec
	// rather than stateless.GenesisCodec
	statelessBlk, err := stateless.Parse(b, stateless.Codec)
	if err != nil {
		return nil, err
	}

	// TODO: remove this to make ParseBlock stateless
	if block, err := vm.GetBlock(statelessBlk.ID()); err == nil {
		// If we have seen this block before, return it with the most up-to-date
		// info
		return block, nil
	}

	return stateful.MakeStateful(
		statelessBlk,
		vm.blkVerifier,
		vm.txExecutorBackend,
		choices.Processing,
	)
}

func (vm *VM) GetBlock(blkID ids.ID) (snowman.Block, error) {
	return vm.blkVerifier.GetStatefulBlock(blkID)
}

// LastAccepted returns the block most recently accepted
func (vm *VM) LastAccepted() (ids.ID, error) {
	return vm.internalState.GetLastAccepted(), nil
}

// SetPreference sets the preferred block to be the one with ID [blkID]
func (vm *VM) SetPreference(blkID ids.ID) error {
	if blkID == vm.preferred {
		// If the preference didn't change, then this is a noop
		return nil
	}
	vm.preferred = blkID
	vm.blockBuilder.ResetBlockTimer()
	return nil
}

func (vm *VM) Preferred() (stateful.Block, error) {
	return vm.blkVerifier.GetStatefulBlock(vm.preferred)
}

func (vm *VM) Version() (string, error) {
	return version.Current.String(), nil
}

// CreateHandlers returns a map where:
// * keys are API endpoint extensions
// * values are API handlers
func (vm *VM) CreateHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	server.RegisterInterceptFunc(vm.metrics.apiRequestMetrics.InterceptRequest)
	server.RegisterAfterFunc(vm.metrics.apiRequestMetrics.AfterRequest)
	if err := server.RegisterService(&Service{vm: vm}, "platform"); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			Handler: server,
		},
	}, nil
}

// CreateStaticHandlers returns a map where:
// * keys are API endpoint extensions
// * values are API handlers
func (vm *VM) CreateStaticHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(&api.StaticService{}, "platform"); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			LockOptions: common.NoLock,
			Handler:     server,
		},
	}, nil
}

func (vm *VM) Connected(vdrID ids.NodeID, _ *version.Application) error {
	return vm.uptimeManager.Connect(vdrID)
}

func (vm *VM) Disconnected(vdrID ids.NodeID) error {
	if err := vm.uptimeManager.Disconnect(vdrID); err != nil {
		return err
	}
	return vm.internalState.Commit()
}

// GetValidatorSet returns the validator set at the specified height for the
// provided subnetID.
func (vm *VM) GetValidatorSet(height uint64, subnetID ids.ID) (map[ids.NodeID]uint64, error) {
	validatorSetsCache, exists := vm.validatorSetCaches[subnetID]
	if !exists {
		validatorSetsCache = &cache.LRU{Size: validatorSetsCacheSize}
		// Only cache whitelisted subnets
		if vm.WhitelistedSubnets.Contains(subnetID) || subnetID == constants.PrimaryNetworkID {
			vm.validatorSetCaches[subnetID] = validatorSetsCache
		}
	}

	if validatorSetIntf, ok := validatorSetsCache.Get(height); ok {
		validatorSet, ok := validatorSetIntf.(map[ids.NodeID]uint64)
		if !ok {
			return nil, errWrongCacheType
		}
		vm.metrics.validatorSetsCached.Inc()
		return validatorSet, nil
	}

	lastAcceptedHeight, err := vm.GetCurrentHeight()
	if err != nil {
		return nil, err
	}
	if lastAcceptedHeight < height {
		return nil, database.ErrNotFound
	}

	// get the start time to track metrics
	startTime := vm.Clock().Time()

	currentValidators, ok := vm.Validators.GetValidators(subnetID)
	if !ok {
		return nil, state.ErrNotEnoughValidators
	}
	currentValidatorList := currentValidators.List()

	vdrSet := make(map[ids.NodeID]uint64, len(currentValidatorList))
	for _, vdr := range currentValidatorList {
		vdrSet[vdr.ID()] = vdr.Weight()
	}

	for i := lastAcceptedHeight; i > height; i-- {
		diffs, err := vm.internalState.GetValidatorWeightDiffs(i, subnetID)
		if err != nil {
			return nil, err
		}

		for nodeID, diff := range diffs {
			var op func(uint64, uint64) (uint64, error)
			if diff.Decrease {
				// The validator's weight was decreased at this block, so in the
				// prior block it was higher.
				op = math.Add64
			} else {
				// The validator's weight was increased at this block, so in the
				// prior block it was lower.
				op = math.Sub64
			}

			newWeight, err := op(vdrSet[nodeID], diff.Amount)
			if err != nil {
				return nil, err
			}
			if newWeight == 0 {
				delete(vdrSet, nodeID)
			} else {
				vdrSet[nodeID] = newWeight
			}
		}
	}

	// cache the validator set
	validatorSetsCache.Put(height, vdrSet)

	endTime := vm.Clock().Time()
	vm.metrics.validatorSetsCreated.Inc()
	vm.metrics.validatorSetsDuration.Add(float64(endTime.Sub(startTime)))
	vm.metrics.validatorSetsHeightDiff.Add(float64(lastAcceptedHeight - height))
	return vdrSet, nil
}

// GetMinimumHeight returns the height of the most recent block beyond the
// horizon of our recentlyAccepted window.
//
// Because the time between blocks is arbitrary, we're only guaranteed that
// the window's configured TTL amount of time has passed once an element
// expires from the window.
//
// To try to always return a block older than the window's TTL, we return the
// parent of the oldest element in the window (as an expired element is always
// guaranteed to be sufficiently stale). If we haven't expired an element yet
// in the case of a process restart, we default to the lastAccepted block's
// height which is likely (but not guaranteed) to also be older than the
// window's configured TTL.
func (vm *VM) GetMinimumHeight() (uint64, error) {
	oldest, ok := vm.recentlyAccepted.Oldest()
	if !ok {
		return vm.GetCurrentHeight()
	}

	blk, err := vm.GetBlock(oldest.(ids.ID))
	if err != nil {
		return 0, err
	}

	return blk.Height() - 1, nil
}

// GetCurrentHeight returns the height of the last accepted block
func (vm *VM) GetCurrentHeight() (uint64, error) {
	lastAccepted, err := vm.GetBlock(vm.internalState.GetLastAccepted())
	if err != nil {
		return 0, err
	}
	return lastAccepted.Height(), nil
}

func (vm *VM) updateValidators() error {
	currentValidators := vm.internalState.CurrentStakers()
	primaryValidators, err := currentValidators.ValidatorSet(constants.PrimaryNetworkID)
	if err != nil {
		return err
	}
	if err := vm.Validators.Set(constants.PrimaryNetworkID, primaryValidators); err != nil {
		return err
	}

	weight, _ := primaryValidators.GetWeight(vm.ctx.NodeID)
	vm.localStake.Set(float64(weight))
	vm.totalStake.Set(float64(primaryValidators.Weight()))

	for subnetID := range vm.WhitelistedSubnets {
		subnetValidators, err := currentValidators.ValidatorSet(subnetID)
		if err != nil {
			return err
		}
		if err := vm.Validators.Set(subnetID, subnetValidators); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VM) CodecRegistry() codec.Registry { return vm.codecRegistry }

func (vm *VM) Clock() *mockable.Clock { return &vm.clock }

func (vm *VM) Logger() logging.Logger { return vm.ctx.Log }

// Returns the percentage of the total stake of the subnet connected to this
// node.
func (vm *VM) getPercentConnected(subnetID ids.ID) (float64, error) {
	vdrSet, exists := vm.Validators.GetValidators(subnetID)
	if !exists {
		return 0, errNoValidators
	}

	vdrSetWeight := vdrSet.Weight()
	if vdrSetWeight == 0 {
		return 1, nil
	}

	var (
		connectedStake uint64
		err            error
	)
	for _, vdr := range vdrSet.List() {
		if !vm.uptimeManager.IsConnected(vdr.ID()) {
			continue // not connected to us --> don't include
		}
		connectedStake, err = math.Add64(connectedStake, vdr.Weight())
		if err != nil {
			return 0, err
		}
	}
	return float64(connectedStake) / float64(vdrSetWeight), nil
}
