// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"fmt"

	"github.com/ava-labs/gecko/vms/components/verify"

	"github.com/ava-labs/gecko/database"
	"github.com/ava-labs/gecko/database/versiondb"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/hashing"
)

// UnsignedAddDefaultSubnetDelegatorTx is an unsigned addDefaultSubnetDelegatorTx
type UnsignedAddDefaultSubnetDelegatorTx struct {
	vm *VM

	// Metadata, inputs and outputs
	CommonTx `serialize:"true"`

	// Describes the delegatee
	DurationValidator `serialize:"true"`

	// Where to send staked AVA after done validating
	Destination ids.ShortID `serialize:"true"`
}

// addDefaultSubnetDelegatorTx is a transaction that, if it is in a
// ProposalBlock that is accepted and followed by a Commit block, adds a
// delegator to the pending validator set of the default subnet. (That is, the
// validator in the tx will have their weight increase at some point in the
// future.) The transaction fee will be paid from the account who signed the
// transaction.
type addDefaultSubnetDelegatorTx struct {
	UnsignedAddDefaultSubnetDelegatorTx `serialize:"true"`

	// Credentials that authorize the inputs to be spent
	Credentials []verify.Verifiable `serialize:"true"`
}

// Creds returns this transactions credentials
func (tx *addDefaultSubnetDelegatorTx) Creds() []verify.Verifiable {
	return tx.Credentials
}

// initialize [tx]
func (tx *addDefaultSubnetDelegatorTx) initialize(vm *VM) error {
	tx.vm = vm
	var err error
	tx.unsignedBytes, err = Codec.Marshal(interface{}(tx.UnsignedAddDefaultSubnetDelegatorTx))
	if err != nil {
		return fmt.Errorf("couldn't marshal UnsignedAddDefaultSubnetDelegatorTx: %w", err)
	}
	tx.bytes, err = Codec.Marshal(tx) // byte representation of the signed transaction
	if err != nil {
		return fmt.Errorf("couldn't marshal addDefaultSubnetDelegatorTx: %w", err)
	}
	tx.id = ids.NewID(hashing.ComputeHash256Array(tx.bytes))
	return nil
}

// SyntacticVerify return nil iff [tx] is valid
// If [tx] is valid, sets [tx.accountID]
// TODO: Only do syntactic Verify once
func (tx *addDefaultSubnetDelegatorTx) SyntacticVerify() error {
	switch {
	case tx == nil:
		return tempError{errNilTx}
	case tx.id.IsZero():
		return tempError{errInvalidID}
	case tx.NetworkID != tx.vm.Ctx.NetworkID:
		return permError{errWrongNetworkID}
	case tx.NodeID.IsZero():
		return tempError{errInvalidID}
	case tx.Wght < MinimumStakeAmount: // Ensure validator is staking at least the minimum amount
		return permError{errWeightTooSmall}
	}

	// Ensure staking length is not too short or long,
	// and that the inputs/outputs of this tx are syntactically valid
	stakingDuration := tx.Duration()
	if stakingDuration < MinimumStakingDuration {
		return permError{errStakeTooShort}
	} else if stakingDuration > MaximumStakingDuration {
		return permError{errStakeTooLong}
	} else if err := syntacticVerifySpend(tx, tx.vm.txFee, tx.vm.avaxAssetID); err != nil {
		return err
	}

	return nil
}

// SemanticVerify this transaction is valid.
// TODO make sure the ins and outs are semantically valid
func (tx *addDefaultSubnetDelegatorTx) SemanticVerify(db database.Database) (*versiondb.Database, *versiondb.Database, func(), func(), TxError) {
	if err := tx.SyntacticVerify(); err != nil {
		return nil, nil, nil, nil, permError{err}
	}

	// Verify inputs/outputs and update the UTXO set
	if err := tx.vm.semanticVerifySpend(db, tx); err != nil {
		return nil, nil, nil, nil, tempError{fmt.Errorf("couldn't verify tx: %w", err)}
	}

	// Ensure the proposed validator starts after the current timestamp
	currentTimestamp, err := tx.vm.getTimestamp(db)
	if err != nil {
		return nil, nil, nil, nil, permError{err}
	}
	validatorStartTime := tx.StartTime()
	if !currentTimestamp.Before(validatorStartTime) {
		return nil, nil, nil, nil, permError{fmt.Errorf("chain timestamp (%s) not before validator's start time (%s)",
			currentTimestamp,
			validatorStartTime)}
	}

	// Ensure that the period this validator validates the specified subnet is a subnet of the time they validate the default subnet
	// First, see if they're currently validating the default subnet
	currentEvents, err := tx.vm.getCurrentValidators(db, DefaultSubnetID)
	if err != nil {
		return nil, nil, nil, nil, permError{fmt.Errorf("couldn't get current validators of default subnet: %v", err)}
	}
	if dsValidator, err := currentEvents.getDefaultSubnetStaker(tx.NodeID); err == nil {
		if !tx.DurationValidator.BoundedBy(dsValidator.StartTime(), dsValidator.EndTime()) {
			return nil, nil, nil, nil, permError{errDSValidatorSubset}
		}
	} else {
		// They aren't currently validating the default subnet.
		// See if they will validate the default subnet in the future.
		pendingDSValidators, err := tx.vm.getPendingValidators(db, DefaultSubnetID)
		if err != nil {
			return nil, nil, nil, nil, permError{fmt.Errorf("couldn't get pending validators of default subnet: %v", err)}
		}
		dsValidator, err := pendingDSValidators.getDefaultSubnetStaker(tx.NodeID)
		if err != nil {
			return nil, nil, nil, nil, permError{errDSValidatorSubset}
		}
		if !tx.DurationValidator.BoundedBy(dsValidator.StartTime(), dsValidator.EndTime()) {
			return nil, nil, nil, nil, permError{errDSValidatorSubset}
		}
	}

	pendingEvents, err := tx.vm.getPendingValidators(db, DefaultSubnetID)
	if err != nil {
		return nil, nil, nil, nil, permError{err}
	}

	pendingEvents.Add(tx) // add validator to set of pending validators

	// If this proposal is committed, update the pending validator set to include the validator,
	// update the validator's account by removing the staked $AVA
	onCommitDB := versiondb.New(db)
	if err := tx.vm.putPendingValidators(onCommitDB, pendingEvents, DefaultSubnetID); err != nil {
		return nil, nil, nil, nil, permError{err}
	}
	/* TODO: Add this (or something like it) back
	if err := tx.vm.putAccount(onCommitDB, newAccount); err != nil {
		return nil, nil, nil, nil, permError{err}
	}
	*/

	// If this proposal is aborted, chain state doesn't change
	onAbortDB := versiondb.New(db)

	return onCommitDB, onAbortDB, nil, nil, nil
}

// InitiallyPrefersCommit returns true if the proposed validators start time is
// after the current wall clock time,
func (tx *addDefaultSubnetDelegatorTx) InitiallyPrefersCommit() bool {
	return tx.StartTime().After(tx.vm.clock.Time())
}

// TODO: Implement
/*
func (vm *VM) newAddDefaultSubnetDelegatorTx(
	nonce,
	weight,
	startTime,
	endTime uint64,
	nodeID ids.ShortID,
	destination ids.ShortID,
	networkID uint32,
	key *crypto.PrivateKeySECP256K1R,
) (*addDefaultSubnetDelegatorTx, error) {
	// Get UTXOs of sender
	addr := key.PublicKey().Address()

	tx := &addDefaultSubnetDelegatorTx{
		UnsignedAddDefaultSubnetDelegatorTx: UnsignedAddDefaultSubnetDelegatorTx{
			DurationValidator: DurationValidator{
				Validator: Validator{
					NodeID: nodeID,
					Wght:   weight,
				},
				Start: startTime,
				End:   endTime,
			},
			NetworkID:   networkID,
			Destination: destination,
		},
	}

	unsignedIntf := interface{}(&tx.UnsignedAddDefaultSubnetDelegatorTx)
	unsignedBytes, err := Codec.Marshal(&unsignedIntf) // byte repr. of unsigned tx
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(unsignedBytes)
	if err != nil {
		return nil, err
	}
	copy(tx.Sig[:], sig)

	return tx, tx.initialize(vm)
}
*/
