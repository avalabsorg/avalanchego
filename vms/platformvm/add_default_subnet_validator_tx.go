// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"

	"github.com/ava-labs/gecko/database"
	"github.com/ava-labs/gecko/database/versiondb"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/constants"
	"github.com/ava-labs/gecko/utils/crypto"
	"github.com/ava-labs/gecko/utils/hashing"
	"github.com/ava-labs/gecko/vms/components/verify"
	"github.com/ava-labs/gecko/vms/secp256k1fx"

	safemath "github.com/ava-labs/gecko/utils/math"
)

var (
	errNilTx          = errors.New("tx is nil")
	errWrongNetworkID = errors.New("tx was issued with a different network ID")
	errWeightTooSmall = errors.New("weight of this validator is too low")
	errStakeTooShort  = errors.New("staking period is too short")
	errStakeTooLong   = errors.New("staking period is too long")
	errTooManyShares  = fmt.Errorf("a staker can only require at most %d shares from delegators", NumberOfShares)
)

// UnsignedAddDefaultSubnetValidatorTx is an unsigned addDefaultSubnetValidatorTx
type UnsignedAddDefaultSubnetValidatorTx struct {
	// Metadata, inputs and outputs
	BaseTx `serialize:"true"`
	// Describes the delegatee
	DurationValidator `serialize:"true"`
	// Where to send staked tokens when done validating
	StakeOwner verify.Verifiable `serialize:"true"`
	// Where to send staking rewards when done validating
	RewardsOwner verify.Verifiable `serialize:"true"`
	// Fee this validator charges delegators as a percentage, times 10,000
	// For example, if this validator has Shares=300,000 then they take 30% of rewards from delegators
	Shares uint32 `serialize:"true"`
}

// initialize [tx]. Sets [tx.vm], [tx.unsignedBytes], [tx.bytes], [tx.id]
func (tx *UnsignedAddDefaultSubnetValidatorTx) initialize(vm *VM, bytes []byte) error {
	if tx.vm != nil { // Already been initialized
		return nil
	}
	tx.vm = vm
	tx.bytes = bytes
	tx.id = ids.NewID(hashing.ComputeHash256Array(bytes))
	var err error
	tx.unsignedBytes, err = Codec.Marshal(interface{}(tx))
	if err != nil {
		return fmt.Errorf("couldn't marshal UnsignedAddDefaultSubnetValidatorTx: %w", err)
	}
	return nil
}

// Verify return nil iff [tx] is valid
func (tx *UnsignedAddDefaultSubnetValidatorTx) Verify() error {
	switch {
	case tx == nil:
		return errNilTx
	case tx.syntacticallyVerified: // already passed syntactic verification
		return nil
	}

	if err := verify.All(
		&tx.BaseTx,
		&tx.DurationValidator,
		tx.StakeOwner,
		tx.RewardsOwner,
	); err != nil {
		return err
	}

	switch {
	case tx.Wght < MinimumStakeAmount: // Ensure validator is staking at least the minimum amount
		return errWeightTooSmall
	case tx.Shares > NumberOfShares: // Ensure delegators shares are in the allowed amount
		return errTooManyShares
	}

	// verify the flow check
	minimumBurnAmount, err := safemath.Add64(tx.vm.txFee, tx.Wght)
	if err != nil {
		return fmt.Errorf("overflow during flow check with fee=%d and weight=%d", tx.vm.txFee, tx.Wght)
	} else if err := syntacticVerifySpend(tx.Ins, tx.Outs, minimumBurnAmount, tx.vm.avaxAssetID); err != nil {
		return err
	}

	// cache that this is valid
	tx.syntacticallyVerified = true
	return nil
}

// SemanticVerify this transaction is valid.
func (tx *UnsignedAddDefaultSubnetValidatorTx) SemanticVerify(
	db database.Database,
	stx *ProposalTx,
) (
	*versiondb.Database,
	*versiondb.Database,
	func() error,
	func() error,
	TxError,
) {
	// Verify the tx is well-formed
	if err := tx.Verify(); err != nil {
		return nil, nil, nil, nil, permError{err}
	}

	// Ensure the proposed validator starts after the current time
	if currentTime, err := tx.vm.getTimestamp(db); err != nil {
		return nil, nil, nil, nil, tempError{err}
	} else if startTime := tx.StartTime(); !currentTime.Before(startTime) {
		return nil, nil, nil, nil, permError{fmt.Errorf("validator's start time (%s) at or after current timestamp (%s)",
			currentTime,
			startTime)}
	}

	// Ensure the proposed validator is not already a validator of the specified subnet
	currentValidators, err := tx.vm.getCurrentValidators(db, constants.DefaultSubnetID)
	if err != nil {
		return nil, nil, nil, nil, tempError{err}
	}
	for _, currentVdr := range tx.vm.getValidators(currentValidators) {
		if currentVdr.ID().Equals(tx.NodeID) {
			return nil, nil, nil, nil, permError{fmt.Errorf("validator %s already is already a Default Subnet validator",
				tx.NodeID)}
		}
	}

	// Ensure the proposed validator is not already slated to validate for the specified subnet
	pendingValidators, err := tx.vm.getPendingValidators(db, constants.DefaultSubnetID)
	if err != nil {
		return nil, nil, nil, nil, tempError{err}
	}
	for _, pendingVdr := range tx.vm.getValidators(pendingValidators) {
		if pendingVdr.ID().Equals(tx.NodeID) {
			return nil, nil, nil, nil, tempError{fmt.Errorf("validator %s is already a pending Default Subnet validator",
				tx.NodeID)}
		}
	}

	// Verify inputs/outputs and update the UTXO set
	onCommitDB := versiondb.New(db)

	// Consume / produce the static UTXOS
	if err := tx.vm.semanticVerifySpend(onCommitDB, tx, tx.Ins, tx.Outs, stx.Credentials); err != nil {
		return nil, nil, nil, nil, err
	}
	// Add validator to set of pending validators
	pendingValidators.Add(stx)
	// If this proposal is committed, update the pending validator set to include the validator
	if err := tx.vm.putPendingValidators(onCommitDB, pendingValidators, constants.DefaultSubnetID); err != nil {
		return nil, nil, nil, nil, tempError{err}
	}

	onAbortDB := versiondb.New(db)

	// Consume / produce the static UTXOS
	if err := tx.vm.semanticVerifySpend(onAbortDB, tx, tx.Ins, tx.Outs, stx.Credentials); err != nil {
		return nil, nil, nil, nil, err
	}

	// Refund the stake here
	txID := tx.ID()
	refundUTXOs, err := tx.vm.generateRefund(
		txID,
		tx.Ins,
		tx.Outs,
		tx.vm.txFee,
		tx.Wght,
		tx.StakeOwner,
	)
	if err != nil {
		return nil, nil, nil, nil, permError{err}
	}

	for _, utxo := range refundUTXOs {
		if err := tx.vm.putUTXO(onAbortDB, utxo); err != nil {
			return nil, nil, nil, nil, tempError{err}
		}
	}

	return onCommitDB, onAbortDB, nil, nil, nil
}

// InitiallyPrefersCommit returns true if the proposed validators start time is
// after the current wall clock time,
func (tx *UnsignedAddDefaultSubnetValidatorTx) InitiallyPrefersCommit() bool {
	return tx.StartTime().After(tx.vm.clock.Time())
}

// NewAddDefaultSubnetValidatorTx returns a new NewAddDefaultSubnetValidatorTx
func (vm *VM) newAddDefaultSubnetValidatorTx(
	stakeAmt, // Amount the delegator stakes
	startTime, // Unix time they start delegating
	endTime uint64, // Unix time they stop delegating
	nodeID ids.ShortID, // ID of the node we are delegating to
	destination ids.ShortID, // Address to returned staked tokens (and maybe reward) to
	shares uint32, // 10,000 times percentage of reward taken from delegators
	keys []*crypto.PrivateKeySECP256K1R, // Keys providing the staked tokens + fee
) (*ProposalTx, error) {
	ins, outs, signers, err := vm.burn(vm.DB, keys, vm.txFee, stakeAmt)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate tx inputs/outputs: %w", err)
	}
	// Create the tx
	utx := &UnsignedAddDefaultSubnetValidatorTx{
		BaseTx: BaseTx{
			NetworkID:    vm.Ctx.NetworkID,
			BlockchainID: vm.Ctx.ChainID,
			Ins:          ins,
			Outs:         outs,
		},
		DurationValidator: DurationValidator{
			Validator: Validator{
				NodeID: nodeID,
				Wght:   stakeAmt,
			},
			Start: startTime,
			End:   endTime,
		},
		StakeOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{destination},
		},
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{destination},
		},
		Shares: shares,
	}
	tx := &ProposalTx{UnsignedProposalTx: utx}
	if err := vm.signProposalTx(tx, signers); err != nil {
		return nil, err
	}
	return tx, utx.Verify()
}
