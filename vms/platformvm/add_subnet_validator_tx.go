// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/timed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
	"github.com/ava-labs/avalanchego/vms/platformvm/validators"
)

var (
	errDSValidatorSubset = errors.New("all subnets' staking period must be a subset of the primary network")

	_ StatefulProposalTx = &StatefulAddSubnetValidatorTx{}
	_ timed.Tx           = &StatefulAddSubnetValidatorTx{}
)

// StatefulAddSubnetValidatorTx is an unsigned addSubnetValidatorTx
type StatefulAddSubnetValidatorTx struct {
	*unsigned.AddSubnetValidatorTx `serialize:"true"`
}

// Attempts to verify this transaction with the provided state.
func (tx *StatefulAddSubnetValidatorTx) SemanticVerify(vm *VM, parentState MutableState, stx *signed.Tx) error {
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
func (tx *StatefulAddSubnetValidatorTx) Execute(
	vm *VM,
	parentState MutableState,
	stx *signed.Tx,
) (
	VersionedState,
	VersionedState,
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
	case len(stx.Creds) == 0:
		return nil, nil, errWrongNumberOfCredentials
	}

	currentStakers := parentState.CurrentStakerChainState()
	pendingStakers := parentState.PendingStakerChainState()

	if vm.bootstrapped.GetValue() {
		currentTimestamp := parentState.GetTimestamp()
		// Ensure the proposed validator starts after the current timestamp
		validatorStartTime := tx.StartTime()
		if !currentTimestamp.Before(validatorStartTime) {
			return nil, nil, fmt.Errorf(
				"validator's start time (%s) is at or after current chain timestamp (%s)",
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

		var vdrTx *unsigned.AddValidatorTx
		if err == nil {
			// This validator is attempting to validate with a currently
			// validing node.
			vdrTx = currentValidator.AddValidatorTx()

			// Ensure that this transaction isn't a duplicate add validator tx.
			subnets := currentValidator.SubnetValidators()
			if _, validates := subnets[tx.Validator.Subnet]; validates {
				return nil, nil, fmt.Errorf(
					"already validating subnet %s",
					tx.Validator.Subnet,
				)
			}
		} else {
			// This validator is attempting to validate with a node that hasn't
			// started validating yet.
			vdrTx, err = pendingStakers.GetValidatorTx(tx.Validator.NodeID)
			if err != nil {
				if err == database.ErrNotFound {
					return nil, nil, errDSValidatorSubset
				}
				return nil, nil, fmt.Errorf(
					"failed to find whether %s is a validator: %w",
					tx.Validator.NodeID,
					err,
				)
			}
		}

		// Ensure that the period this validator validates the specified subnet
		// is a subset of the time they validate the primary network.
		if !tx.Validator.BoundedBy(vdrTx.StartTime(), vdrTx.EndTime()) {
			return nil, nil, errDSValidatorSubset
		}

		// Ensure that this transaction isn't a duplicate add validator tx.
		pendingValidator := pendingStakers.GetValidator(tx.Validator.NodeID)
		subnets := pendingValidator.SubnetValidators()
		if _, validates := subnets[tx.Validator.Subnet]; validates {
			return nil, nil, fmt.Errorf(
				"already validating subnet %s",
				tx.Validator.Subnet,
			)
		}

		baseTxCredsLen := len(stx.Creds) - 1
		baseTxCreds := stx.Creds[:baseTxCredsLen]
		subnetCred := stx.Creds[baseTxCredsLen]

		subnetIntf, _, err := parentState.GetTx(tx.Validator.Subnet)
		if err != nil {
			if err == database.ErrNotFound {
				return nil, nil, errDSValidatorSubset
			}
			return nil, nil, fmt.Errorf(
				"couldn't find subnet %s with %w",
				tx.Validator.Subnet,
				err,
			)
		}

		subnet, ok := subnetIntf.Unsigned.(*unsigned.CreateSubnetTx)
		if !ok {
			return nil, nil, fmt.Errorf(
				"%s is not a subnet",
				tx.Validator.Subnet,
			)
		}

		if err := vm.fx.VerifyPermission(tx, tx.SubnetAuth, subnetCred, subnet.Owner); err != nil {
			return nil, nil, err
		}

		// Verify the flowcheck
		if err := vm.semanticVerifySpend(parentState, tx, tx.Ins, tx.Outs, baseTxCreds, vm.TxFee, vm.ctx.AVAXAssetID); err != nil {
			return nil, nil, err
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
	onCommitState := newVersionedState(parentState, currentStakers, newlyPendingStakers)

	// Consume the UTXOS
	consumeInputs(onCommitState, tx.Ins)
	// Produce the UTXOS
	txID := tx.ID()
	produceOutputs(onCommitState, txID, vm.ctx.AVAXAssetID, tx.Outs)

	// Set up the state if this tx is aborted
	onAbortState := newVersionedState(parentState, currentStakers, pendingStakers)
	// Consume the UTXOS
	consumeInputs(onAbortState, tx.Ins)
	// Produce the UTXOS
	produceOutputs(onAbortState, txID, vm.ctx.AVAXAssetID, tx.Outs)

	return onCommitState, onAbortState, nil
}

// InitiallyPrefersCommit returns true if the proposed validators start time is
// after the current wall clock time,
func (tx *StatefulAddSubnetValidatorTx) InitiallyPrefersCommit(vm *VM) bool {
	return tx.StartTime().After(vm.clock.Time())
}

// Create a new transaction
func (vm *VM) newAddSubnetValidatorTx(
	weight, // Sampling weight of the new validator
	startTime, // Unix time they start delegating
	endTime uint64, // Unix time they top delegating
	nodeID ids.NodeID, // ID of the node validating
	subnetID ids.ID, // ID of the subnet the validator will validate
	keys []*crypto.PrivateKeySECP256K1R, // Keys to use for adding the validator
	changeAddr ids.ShortID, // Address to send change to, if there is any
) (*signed.Tx, error) {
	ins, outs, _, signers, err := vm.stake(keys, 0, vm.TxFee, changeAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate tx inputs/outputs: %w", err)
	}

	subnetAuth, subnetSigners, err := vm.authorize(vm.internalState, subnetID, keys)
	if err != nil {
		return nil, fmt.Errorf("couldn't authorize tx's subnet restrictions: %w", err)
	}
	signers = append(signers, subnetSigners)

	// Create the tx
	utx := &unsigned.AddSubnetValidatorTx{
		BaseTx: unsigned.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    vm.ctx.NetworkID,
			BlockchainID: vm.ctx.ChainID,
			Ins:          ins,
			Outs:         outs,
		}},
		Validator: validators.SubnetValidator{
			Validator: validators.Validator{
				NodeID: nodeID,
				Start:  startTime,
				End:    endTime,
				Wght:   weight,
			},
			Subnet: subnetID,
		},
		SubnetAuth: subnetAuth,
	}
	tx := &signed.Tx{Unsigned: utx}
	if err := tx.Sign(Codec, signers); err != nil {
		return nil, err
	}
	return tx, utx.SyntacticVerify(vm.ctx)
}
