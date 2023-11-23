// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/snow/uptime"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/vms/platformvm/reward"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ txs.Visitor = (*ProposalTxPreference)(nil)

	ErrMissingStakerTx                      = errors.New("failed to get staker transaction")
	ErrUnexpectedStakerTransactionType      = errors.New("unexpected staker transaction type")
	ErrStakerWithoutPrimaryNetworkValidator = errors.New("failed to get primary network validator of staker")
	ErrMissingSubnetTransformation          = errors.New("failed to get subnet transformation")
	ErrCalculatingUptime                    = errors.New("failed calculating uptime")
)

type ProposalTxPreference struct {
	// inputs, to be filled before visitor methods are called
	PrimaryUptimePercentage float64
	Uptimes                 uptime.Calculator
	State                   state.Chain
	Tx                      *txs.Tx

	// outputs populated by this struct's methods:
	//
	// [PrefersCommit] is true iff this node initially prefers to commit the
	// block containing this transaction.
	PrefersCommit bool
}

func (*ProposalTxPreference) CreateChainTx(*txs.CreateChainTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) CreateSubnetTx(*txs.CreateSubnetTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) ImportTx(*txs.ImportTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) ExportTx(*txs.ExportTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) RemoveSubnetValidatorTx(*txs.RemoveSubnetValidatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) TransformSubnetTx(*txs.TransformSubnetTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AddPermissionlessValidatorTx(*txs.AddPermissionlessValidatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AddPermissionlessDelegatorTx(*txs.AddPermissionlessDelegatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) TransferSubnetOwnershipTx(*txs.TransferSubnetOwnershipTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) BaseTx(*txs.BaseTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AddValidatorTx(*txs.AddValidatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AddSubnetValidatorTx(*txs.AddSubnetValidatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AddDelegatorTx(*txs.AddDelegatorTx) error {
	return ErrWrongTxType
}

func (*ProposalTxPreference) AdvanceTimeTx(*txs.AdvanceTimeTx) error {
	return ErrWrongTxType
}

func (e *ProposalTxPreference) RewardValidatorTx(tx *txs.RewardValidatorTx) error {
	stakerTx, _, err := e.State.GetTx(tx.TxID)
	if err == database.ErrNotFound {
		// This can happen if this transaction is attempting to reward a
		// validator that hasn't been confirmed.
		e.PrefersCommit = true
		return nil
	}
	if err != nil {
		// GetTx can only return [ErrNotFound], or an unexpected error like a
		// parsing error or internal DB error. For anything other than
		// [ErrNotFound] the block can just be dropped.
		return fmt.Errorf("%w %s: %w",
			ErrMissingStakerTx,
			tx.TxID,
			err,
		)
	}

	staker, ok := stakerTx.Unsigned.(txs.Staker)
	if !ok {
		// Because this transaction isn't guaranteed to have been verified yet,
		// it's possible that a malicious node issued this transaction into a
		// block that will fail verification in the future.
		return fmt.Errorf("%w %s: %T",
			ErrUnexpectedStakerTransactionType,
			tx.TxID,
			stakerTx.Unsigned,
		)
	}

	// retrieve primaryNetworkValidator before possibly removing it.
	nodeID := staker.NodeID()
	primaryNetworkValidator, err := e.State.GetCurrentValidator(
		constants.PrimaryNetworkID,
		nodeID,
	)
	if err != nil {
		// If this transaction is included into an invalid block where the
		// staker has already been removed, we can just drop it.
		return fmt.Errorf("%w %s: %w",
			ErrStakerWithoutPrimaryNetworkValidator,
			nodeID,
			err,
		)
	}

	expectedUptimePercentage := e.PrimaryUptimePercentage
	if subnetID := staker.SubnetID(); subnetID != constants.PrimaryNetworkID {
		transformSubnet, err := GetTransformSubnetTx(e.State, subnetID)
		if err != nil {
			// If the subnet hasn't been transformed yet, the tx we are removing
			// isn't for a permissionless subnet. So, it's removal here is
			// invalid.
			return fmt.Errorf("%w %s: %w",
				ErrMissingSubnetTransformation,
				subnetID,
				err,
			)
		}

		expectedUptimePercentage = float64(transformSubnet.UptimeRequirement) / reward.PercentDenominator
	}

	uptime, err := e.Uptimes.CalculateUptimePercentFrom(
		nodeID,
		constants.PrimaryNetworkID,
		primaryNetworkValidator.StartTime,
	)
	if err != nil {
		// If this transaction is included into an invalid block where the
		// staker has already been removed, we can just drop it.
		return fmt.Errorf("%w: %w",
			ErrCalculatingUptime,
			nodeID,
			err,
		)
	}

	e.PrefersCommit = uptime >= expectedUptimePercentage
	return nil
}
