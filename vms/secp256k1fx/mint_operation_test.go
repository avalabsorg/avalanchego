// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package secp256k1fx

import (
	"testing"

	"github.com/ava-labs/gecko/ids"
)

func TestMintOperationVerifyNil(t *testing.T) {
	op := (*MintOperation)(nil)
	if err := op.Verify(); err == nil {
		t.Fatalf("MintOperation.Verify should have returned an error due to an nil operation")
	}
}

func TestMintOperationOuts(t *testing.T) {
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					ids.NewShortID(addrBytes),
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt:      1,
			Locktime: 0,
			OutputOwners: OutputOwners{
				Threshold: 1,
			},
		},
	}

	outs := op.Outs()
	if len(outs) != 2 {
		t.Fatalf("Wrong number of outputs")
	}
}
