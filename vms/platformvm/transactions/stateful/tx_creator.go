// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
)

func MakeStatefulTx(tx *signed.Tx) (Tx, error) {
	switch utx := tx.Unsigned.(type) {
	case *unsigned.AddDelegatorTx:
		return &AddDelegatorTx{
			AddDelegatorTx: utx,
			txID:           tx.ID(),
			signedBytes:    tx.Bytes(),
		}, nil

	case *unsigned.AddSubnetValidatorTx:
		return &AddSubnetValidatorTx{
			AddSubnetValidatorTx: utx,
			txID:                 tx.ID(),
			signedBytes:          tx.Bytes(),
		}, nil

	case *unsigned.AddValidatorTx:
		return &AddValidatorTx{
			AddValidatorTx: utx,
			txID:           tx.ID(),
			signedBytes:    tx.Bytes(),
		}, nil
	case *unsigned.AdvanceTimeTx:
		return &AdvanceTimeTx{
			AdvanceTimeTx: utx,
			ID:            tx.ID(),
		}, nil
	case *unsigned.RewardValidatorTx:
		return &RewardValidatorTx{
			RewardValidatorTx: utx,
			ID:                tx.ID(),
		}, nil
	case *unsigned.CreateChainTx:
		return &CreateChainTx{
			CreateChainTx: utx,
			txID:          tx.ID(),
			signedBytes:   tx.Bytes(),
		}, nil
	case *unsigned.CreateSubnetTx:
		return &CreateSubnetTx{
			CreateSubnetTx: utx,
			txID:           tx.ID(),
			signedBytes:    tx.Bytes(),
		}, nil
	case *unsigned.ImportTx:
		return &ImportTx{
			ImportTx: utx,
			txID:     tx.ID(),
		}, nil
	case *unsigned.ExportTx:
		return &ExportTx{
			ExportTx:    utx,
			txID:        tx.ID(),
			signedBytes: tx.Bytes(),
		}, nil

	default:
		return nil, fmt.Errorf("unverifiable tx type")
	}
}
