// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/api"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	pChainValidator "github.com/ava-labs/avalanchego/vms/platformvm/validator"
)

var (
	_ StatefulProposalTx = &StatefulAddDelegatorTx{}

	errDelegatorSubset = errors.New("delegator's time range must be a subset of the validator's time range")
	errInvalidState    = errors.New("generated output isn't valid state")
	errOverDelegated   = errors.New("validator would be over delegated")
)

// StatefulAddDelegatorTx is an unsigned addDelegatorTx
type StatefulAddDelegatorTx struct {
	*unsigned.AddDelegatorTx `serialize:"true"`
}

// Attempts to verify this transaction with the provided state.
func (tx *StatefulAddDelegatorTx) SemanticVerify(vm *VM, parentState state.Mutable, stx *signed.Tx) error {
	startTime := tx.StartTime()
	maxLocalStartTime := vm.clock.Time().Add(maxFutureStartTime)
	if startTime.After(maxLocalStartTime) {
		return errFutureStakeTime
	}

	_, _, err := tx.Execute(vm, parentState, stx)
	// We ignore [errFutureStakeTime] here because an advanceTimeTx will be
	// issued before this transaction is issued.
	if errors.Is(err, errFutureStakeTime) {
		return nil
	}
	return err
}

// Execute this transaction.
func (tx *StatefulAddDelegatorTx) Execute(
	vm *VM,
	parentState state.Mutable,
	stx *signed.Tx,
) (
	state.Versioned,
	state.Versioned,
	error,
) {
	// Verify the tx is well-formed
	if err := tx.SyntacticVerify(vm.ctx); err != nil {
		return nil, nil, err
	}

	duration := tx.Validator.Duration()
	switch {
	case duration < vm.MinStakeDuration: // Ensure staking length is not too short
		return nil, nil, errStakeTooShort
	case duration > vm.MaxStakeDuration: // Ensure staking length is not too long
		return nil, nil, errStakeTooLong
	case tx.Validator.Wght < vm.MinDelegatorStake:
		// Ensure validator is staking at least the minimum amount
		return nil, nil, errWeightTooSmall
	}

	outs := make([]*avax.TransferableOutput, len(tx.Outs)+len(tx.Stake))
	copy(outs, tx.Outs)
	copy(outs[len(tx.Outs):], tx.Stake)

	currentStakers := parentState.CurrentStakerChainState()
	pendingStakers := parentState.PendingStakerChainState()

	if vm.bootstrapped.GetValue() {
		currentTimestamp := parentState.GetTimestamp()
		// Ensure the proposed validator starts after the current timestamp
		validatorStartTime := tx.StartTime()
		if !currentTimestamp.Before(validatorStartTime) {
			return nil, nil, fmt.Errorf(
				"chain timestamp (%s) not before validator's start time (%s)",
				currentTimestamp,
				validatorStartTime,
			)
		}

		currentValidator, err := currentStakers.GetValidator(tx.Validator.NodeID)
		if err != nil && err != database.ErrNotFound {
			return nil, nil, fmt.Errorf(
				"failed to find whether %s is a validator: %w",
				tx.Validator.NodeID,
				err,
			)
		}

		pendingValidator := pendingStakers.GetValidator(tx.Validator.NodeID)
		pendingDelegators := pendingValidator.Delegators()

		var (
			vdrTx                  *unsigned.AddValidatorTx
			currentDelegatorWeight uint64
			currentDelegators      []*unsigned.AddDelegatorTx
		)
		if err == nil {
			// This delegator is attempting to delegate to a currently validing
			// node.
			vdrTx = currentValidator.AddValidatorTx()
			currentDelegatorWeight = currentValidator.DelegatorWeight()
			currentDelegators = currentValidator.Delegators()
		} else {
			// This delegator is attempting to delegate to a node that hasn't
			// started validating yet.
			vdrTx, err = pendingStakers.GetValidatorTx(tx.Validator.NodeID)
			if err != nil {
				if err == database.ErrNotFound {
					return nil, nil, errDelegatorSubset
				}
				return nil, nil, fmt.Errorf(
					"failed to find whether %s is a validator: %w",
					tx.Validator.NodeID,
					err,
				)
			}
		}

		// Ensure that the period this delegator delegates is a subset of the
		// time the validator validates.
		if !tx.Validator.BoundedBy(vdrTx.StartTime(), vdrTx.EndTime()) {
			return nil, nil, errDelegatorSubset
		}

		// Ensure that the period this delegator delegates wouldn't become over
		// delegated.
		vdrWeight := vdrTx.Weight()
		currentWeight, err := math.Add64(vdrWeight, currentDelegatorWeight)
		if err != nil {
			return nil, nil, err
		}

		maximumWeight, err := math.Mul64(MaxValidatorWeightFactor, vdrWeight)
		if err != nil {
			return nil, nil, api.ErrStakeOverflow
		}

		if !currentTimestamp.Before(vm.ApricotPhase3Time) {
			maximumWeight = math.Min64(maximumWeight, vm.MaxValidatorStake)
		}

		canDelegate, err := CanDelegate(
			currentDelegators,
			pendingDelegators,
			tx.AddDelegatorTx,
			currentWeight,
			maximumWeight,
		)
		if err != nil {
			return nil, nil, err
		}
		if !canDelegate {
			return nil, nil, errOverDelegated
		}

		// Verify the flowcheck
		if err := vm.semanticVerifySpend(
			parentState,
			tx.AddDelegatorTx,
			tx.Ins,
			outs,
			stx.Creds,
			vm.AddStakerTxFee,
			vm.ctx.AVAXAssetID,
		); err != nil {
			return nil, nil, fmt.Errorf("failed semanticVerifySpend: %w", err)
		}

		// Make sure the tx doesn't start too far in the future. This is done
		// last to allow SemanticVerification to explicitly check for this
		// error.
		maxStartTime := currentTimestamp.Add(maxFutureStartTime)
		if validatorStartTime.After(maxStartTime) {
			return nil, nil, errFutureStakeTime
		}
	}

	// Set up the state if this tx is committed
	newlyPendingStakers := pendingStakers.AddStaker(stx)
	onCommitState := state.NewVersioned(parentState, currentStakers, newlyPendingStakers)

	// Consume the UTXOS
	consumeInputs(onCommitState, tx.Ins)
	// Produce the UTXOS
	txID := tx.ID()
	produceOutputs(onCommitState, txID, vm.ctx.AVAXAssetID, tx.Outs)

	// Set up the state if this tx is aborted
	onAbortState := state.NewVersioned(parentState, currentStakers, pendingStakers)
	// Consume the UTXOS
	consumeInputs(onAbortState, tx.Ins)
	// Produce the UTXOS
	produceOutputs(onAbortState, txID, vm.ctx.AVAXAssetID, outs)

	return onCommitState, onAbortState, nil
}

// InitiallyPrefersCommit returns true if the proposed validators start time is
// after the current wall clock time,
func (tx *StatefulAddDelegatorTx) InitiallyPrefersCommit(vm *VM) bool {
	return tx.StartTime().After(vm.clock.Time())
}

// Creates a new transaction
func (vm *VM) newAddDelegatorTx(
	stakeAmt, // Amount the delegator stakes
	startTime, // Unix time they start delegating
	endTime uint64, // Unix time they stop delegating
	nodeID ids.NodeID, // ID of the node we are delegating to
	rewardAddress ids.ShortID, // Address to send reward to, if applicable
	keys []*crypto.PrivateKeySECP256K1R, // Keys providing the staked tokens
	changeAddr ids.ShortID, // Address to send change to, if there is any
) (*signed.Tx, error) {
	ins, unlockedOuts, lockedOuts, signers, err := vm.stake(keys, stakeAmt, vm.AddStakerTxFee, changeAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate tx inputs/outputs: %w", err)
	}
	// Create the tx
	utx := &unsigned.AddDelegatorTx{
		BaseTx: unsigned.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    vm.ctx.NetworkID,
			BlockchainID: vm.ctx.ChainID,
			Ins:          ins,
			Outs:         unlockedOuts,
		}},
		Validator: pChainValidator.Validator{
			NodeID: nodeID,
			Start:  startTime,
			End:    endTime,
			Wght:   stakeAmt,
		},
		Stake: lockedOuts,
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{rewardAddress},
		},
	}
	tx := &signed.Tx{Unsigned: utx}
	if err := tx.Sign(Codec, signers); err != nil {
		return nil, err
	}
	return tx, utx.SyntacticVerify(vm.ctx)
}

// CanDelegate returns if the [new] delegator can be added to a validator who
// has [current] and [pending] delegators. [currentStake] is the current amount
// of stake on the validator, include the [current] delegators. [maximumStake]
// is the maximum amount of stake that can be on the validator at any given
// time. It is assumed that the validator without adding [new] does not violate
// [maximumStake].
func CanDelegate(
	current,
	pending []*unsigned.AddDelegatorTx, // sorted by next start time first
	new *unsigned.AddDelegatorTx,
	currentStake,
	maximumStake uint64,
) (bool, error) {
	maxStake, err := maxStakeAmount(current, pending, new.StartTime(), new.EndTime(), currentStake)
	if err != nil {
		return false, err
	}
	newMaxStake, err := math.Add64(maxStake, new.Validator.Wght)
	if err != nil {
		return false, err
	}
	return newMaxStake <= maximumStake, nil
}

// Return the maximum amount of stake on a node (including delegations) at any
// given time between [startTime] and [endTime] given that:
// * The amount of stake on the node right now is [currentStake]
// * The delegations currently on this node are [current]
// * [current] is sorted in order of increasing delegation end time.
// * The stake delegated in [current] are already included in [currentStake]
// * [startTime] is in the future, and [endTime] > [startTime]
// * The delegations that will be on this node in the future are [pending]
// * The start time of all delegations in [pending] are in the future
// * [pending] is sorted in order of increasing delegation start time
func maxStakeAmount(
	current,
	pending []*unsigned.AddDelegatorTx, // sorted by next start time first
	startTime time.Time,
	endTime time.Time,
	currentStake uint64,
) (uint64, error) {
	// Keep track of which delegators should be removed next so that we can
	// efficiently remove delegators and keep the current stake updated.
	toRemoveHeap := pChainValidator.EndTimeHeap{}
	for _, currentDelegator := range current {
		toRemoveHeap.Add(&currentDelegator.Validator)
	}

	var (
		err error
		// [maxStake] is the max stake at any point between now [starTime] and [endTime]
		maxStake uint64
	)

	// Calculate what the amount staked will be when each pending delegation
	// starts.
	for _, nextPending := range pending { // Iterates in order of increasing start time
		// Calculate what the amount staked will be when this delegation starts.
		nextPendingStartTime := nextPending.StartTime()

		if nextPendingStartTime.After(endTime) {
			// This delegation starts after [endTime].
			// Since we're calculating the max amount staked in
			// [startTime, endTime], we can stop. (Recall that [pending] is
			// sorted in order of increasing end time.)
			break
		}

		// Subtract from [currentStake] all of the current delegations that will
		// have ended by the time that the delegation [nextPending] starts.
		for toRemoveHeap.Len() > 0 {
			// Get the next current delegation that will end.
			toRemove := toRemoveHeap.Peek()
			toRemoveEndTime := toRemove.EndTime()
			if toRemoveEndTime.After(nextPendingStartTime) {
				break
			}
			// This current delegation [toRemove] ends before [nextPending]
			// starts, so its stake should be subtracted from [currentStake].

			// Changed in AP3:
			// If the new delegator has started, then this current delegator
			// should have an end time that is > [startTime].
			newDelegatorHasStartedBeforeFinish := toRemoveEndTime.After(startTime)
			if newDelegatorHasStartedBeforeFinish && currentStake > maxStake {
				// Only update [maxStake] if it's after [startTime]
				maxStake = currentStake
			}

			currentStake, err = math.Sub64(currentStake, toRemove.Wght)
			if err != nil {
				return 0, err
			}

			// Changed in AP3:
			// Remove the delegator from the heap and update the heap so that
			// the top of the heap is the next delegator to remove.
			toRemoveHeap.Remove()
		}

		// Add to [currentStake] the stake of this pending delegator to
		// calculate what the stake will be when this pending delegation has
		// started.
		currentStake, err = math.Add64(currentStake, nextPending.Validator.Wght)
		if err != nil {
			return 0, err
		}

		// Changed in AP3:
		// If the new delegator has started, then this pending delegator should
		// have a start time that is >= [startTime]. Otherwise, the delegator
		// hasn't started yet and the [currentStake] shouldn't count towards the
		// [maximumStake] during the delegators delegation period.
		newDelegatorHasStarted := !nextPendingStartTime.Before(startTime)
		if newDelegatorHasStarted && currentStake > maxStake {
			// Only update [maxStake] if it's after [startTime]
			maxStake = currentStake
		}

		// This pending delegator is a current delegator relative
		// when considering later pending delegators that start late
		toRemoveHeap.Add(&nextPending.Validator)
	}

	// [currentStake] is now the amount staked before the next pending delegator
	// whose start time is after [endTime].

	// If there aren't any delegators that will be added before the end of our
	// delegation period, we should advance through time until our delegation
	// period starts.
	for toRemoveHeap.Len() > 0 {
		toRemove := toRemoveHeap.Peek()
		toRemoveEndTime := toRemove.EndTime()
		if toRemoveEndTime.After(startTime) {
			break
		}

		currentStake, err = math.Sub64(currentStake, toRemove.Wght)
		if err != nil {
			return 0, err
		}

		// Changed in AP3:
		// Remove the delegator from the heap and update the heap so that the
		// top of the heap is the next delegator to remove.
		toRemoveHeap.Remove()
	}

	// We have advanced time to be inside the delegation window.
	// Make sure that the max stake is updated accordingly.
	if currentStake > maxStake {
		maxStake = currentStake
	}
	return maxStake, nil
}

func (vm *VM) maxStakeAmount(
	subnetID ids.ID,
	nodeID ids.NodeID,
	startTime time.Time,
	endTime time.Time,
) (uint64, error) {
	if startTime.After(endTime) {
		return 0, errStartAfterEndTime
	}
	if timestamp := vm.internalState.GetTimestamp(); startTime.Before(timestamp) {
		return 0, errStartTimeTooEarly
	}
	if subnetID == constants.PrimaryNetworkID {
		return vm.maxPrimarySubnetStakeAmount(nodeID, startTime, endTime)
	}
	return vm.maxSubnetStakeAmount(subnetID, nodeID, startTime, endTime)
}

func (vm *VM) maxSubnetStakeAmount(
	subnetID ids.ID,
	nodeID ids.NodeID,
	startTime time.Time,
	endTime time.Time,
) (uint64, error) {
	var (
		vdrTx  *unsigned.AddSubnetValidatorTx
		exists bool
	)

	pendingStakers := vm.internalState.PendingStakerChainState()
	pendingValidator := pendingStakers.GetValidator(nodeID)

	currentStakers := vm.internalState.CurrentStakerChainState()
	currentValidator, err := currentStakers.GetValidator(nodeID)
	switch err {
	case nil:
		vdrTx, exists = currentValidator.SubnetValidators()[subnetID]
		if !exists {
			vdrTx = pendingValidator.SubnetValidators()[subnetID]
		}
	case database.ErrNotFound:
		vdrTx = pendingValidator.SubnetValidators()[subnetID]
	default:
		return 0, err
	}

	if vdrTx == nil {
		return 0, nil
	}
	if vdrTx.StartTime().After(endTime) {
		return 0, nil
	}
	if vdrTx.EndTime().Before(startTime) {
		return 0, nil
	}
	return vdrTx.Weight(), nil
}

func (vm *VM) maxPrimarySubnetStakeAmount(
	nodeID ids.NodeID,
	startTime time.Time,
	endTime time.Time,
) (uint64, error) {
	currentStakers := vm.internalState.CurrentStakerChainState()
	pendingStakers := vm.internalState.PendingStakerChainState()

	pendingValidator := pendingStakers.GetValidator(nodeID)
	currentValidator, err := currentStakers.GetValidator(nodeID)

	switch err {
	case nil:
		vdrTx := currentValidator.AddValidatorTx()
		if vdrTx.StartTime().After(endTime) {
			return 0, nil
		}
		if vdrTx.EndTime().Before(startTime) {
			return 0, nil
		}

		currentWeight := vdrTx.Weight()
		currentWeight, err = math.Add64(currentWeight, currentValidator.DelegatorWeight())
		if err != nil {
			return 0, err
		}
		return maxStakeAmount(
			currentValidator.Delegators(),
			pendingValidator.Delegators(),
			startTime,
			endTime,
			currentWeight,
		)
	case database.ErrNotFound:
		futureValidator, err := pendingStakers.GetValidatorTx(nodeID)
		if err == database.ErrNotFound {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		if futureValidator.StartTime().After(endTime) {
			return 0, nil
		}
		if futureValidator.EndTime().Before(startTime) {
			return 0, nil
		}

		return maxStakeAmount(
			nil,
			pendingValidator.Delegators(),
			startTime,
			endTime,
			futureValidator.Weight(),
		)
	default:
		return 0, err
	}
}
