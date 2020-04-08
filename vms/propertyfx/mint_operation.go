package propertyfx

import (
	"errors"

	"github.com/ava-labs/gecko/vms/components/verify"
	"github.com/ava-labs/gecko/vms/secp256k1fx"
)

var (
	errNilMintOperation = errors.New("nil mint operation")
)

// MintOperation ...
type MintOperation struct {
	MintInput   secp256k1fx.Input `serialize:"true" json:"mintInput"`
	MintOutput  MintOutput        `serialize:"true" json:"mintOutput"`
	OwnedOutput OwnedOutput       `serialize:"true" json:"ownedOutput"`
}

// Outs ...
func (op *MintOperation) Outs() []verify.Verifiable {
	return []verify.Verifiable{
		&op.MintOutput,
		&op.OwnedOutput,
	}
}

// Verify ...
func (op *MintOperation) Verify() error {
	switch {
	case op == nil:
		return errNilMintOperation
	default:
		return verify.All(&op.MintInput, &op.MintOutput, &op.OwnedOutput)
	}
}
