// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/reward"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/stretchr/testify/assert"
)

func TestAllSignedTxMarshalling(t *testing.T) {
	h := newTestHelpersCollection()
	defer func() {
		if err := internalStateShutdown(h); err != nil {
			t.Fatal(err)
		}
	}()

	type test struct {
		name           string
		createSignedTx func() (*signed.Tx, error)
		unsignedBytes  []byte
		signedBytes    []byte
	}

	tests := []test{
		{
			name: "AddValidator",
			createSignedTx: func() (*signed.Tx, error) {
				rewardAddress := preFundedKeys[0].PublicKey().Address()
				nodeID := ids.NodeID(rewardAddress)

				return h.txBuilder.NewAddValidatorTx(
					h.cfg.MinValidatorStake,
					uint64(defaultValidateStartTime.Unix()),
					uint64(defaultValidateEndTime.Unix()),
					nodeID,
					rewardAddress,
					reward.PercentDenominator,
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
					ids.ShortEmpty,
				)
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0x81, 0x19, 0xc0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x0, 0x32, 0xc9, 0xa9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0xd6, 0xd8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x4b, 0x40, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x4b, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0xf, 0x42, 0x40},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0x81, 0x19, 0xc0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x0, 0x32, 0xc9, 0xa9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0xd6, 0xd8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x4b, 0x40, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x4b, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0xf, 0x42, 0x40, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x1, 0x11, 0xe3, 0x86, 0x4f, 0xf, 0x95, 0xf1, 0x7d, 0xe6, 0xd9, 0x44, 0xfd, 0xc8, 0x55, 0x9d, 0xbb, 0x1, 0x3e, 0x17, 0x33, 0xf7, 0x8b, 0x64, 0x5b, 0xb9, 0xf9, 0x37, 0xad, 0xa1, 0x28, 0x43, 0xa0, 0x2d, 0x1b, 0x44, 0x7b, 0xb7, 0x99, 0xc7, 0x3e, 0x12, 0x32, 0xab, 0x7f, 0xe9, 0x71, 0x63, 0x9a, 0x7a, 0xa3, 0xaa, 0xd7, 0x9d, 0x4b, 0xa4, 0x86, 0x61, 0xee, 0xf6, 0x55, 0x3e, 0xd2, 0x49, 0x72, 0x0},
		},
		{
			name: "AddDelegator",
			createSignedTx: func() (*signed.Tx, error) {
				rewardAddress := preFundedKeys[0].PublicKey().Address()
				nodeID := ids.NodeID(rewardAddress)

				return h.txBuilder.NewAddDelegatorTx(
					h.cfg.MinDelegatorStake,
					uint64(defaultValidateStartTime.Unix()),
					uint64(defaultValidateEndTime.Unix()),
					nodeID,
					rewardAddress,
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
					ids.ShortEmpty,
				)
			},
			unsignedBytes: []byte{0, 0, 0, 0, 0, 14, 0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 0, 0, 29, 190, 34, 192, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 29, 205, 101, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 252, 237, 168, 249, 15, 203, 93, 48, 97, 75, 153, 215, 159, 196, 186, 162, 147, 7, 118, 38, 0, 0, 0, 0, 50, 201, 169, 0, 0, 0, 0, 0, 50, 214, 216, 0, 0, 0, 0, 0, 0, 15, 66, 64, 0, 0, 0, 1, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 15, 66, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 252, 237, 168, 249, 15, 203, 93, 48, 97, 75, 153, 215, 159, 196, 186, 162, 147, 7, 118, 38},
			signedBytes:   []byte{0, 0, 0, 0, 0, 14, 0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 0, 0, 29, 190, 34, 192, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 29, 205, 101, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 252, 237, 168, 249, 15, 203, 93, 48, 97, 75, 153, 215, 159, 196, 186, 162, 147, 7, 118, 38, 0, 0, 0, 0, 50, 201, 169, 0, 0, 0, 0, 0, 50, 214, 216, 0, 0, 0, 0, 0, 0, 15, 66, 64, 0, 0, 0, 1, 121, 101, 101, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 15, 66, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 252, 237, 168, 249, 15, 203, 93, 48, 97, 75, 153, 215, 159, 196, 186, 162, 147, 7, 118, 38, 0, 0, 0, 1, 0, 0, 0, 9, 0, 0, 0, 1, 202, 225, 157, 242, 253, 127, 37, 89, 87, 134, 232, 129, 86, 75, 169, 104, 4, 167, 255, 72, 189, 163, 94, 155, 128, 45, 182, 36, 37, 222, 195, 104, 4, 67, 34, 148, 74, 149, 133, 161, 61, 8, 101, 240, 253, 126, 47, 38, 1, 28, 127, 202, 145, 94, 65, 183, 17, 147, 134, 106, 137, 154, 77, 143, 0},
		},
		{
			name: "AddSubnetValidator",
			createSignedTx: func() (*signed.Tx, error) {
				nodeID := ids.NodeID(preFundedKeys[0].PublicKey().Address())

				return h.txBuilder.NewAddSubnetValidatorTx(
					defaultWeight,
					uint64(defaultValidateStartTime.Unix()),
					uint64(defaultValidateEndTime.Unix()),
					nodeID,
					testSubnet1.ID(),
					[]*crypto.PrivateKeySECP256K1R{testSubnet1ControlKeys[0], testSubnet1ControlKeys[1]},
					ids.ShortEmpty,
				)
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xd, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x64, 0x9c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x0, 0x32, 0xc9, 0xa9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0xd6, 0xd8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x27, 0x10, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xd, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x64, 0x9c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x0, 0x32, 0xc9, 0xa9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0xd6, 0xd8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x27, 0x10, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x1, 0x60, 0xfc, 0x3f, 0x2a, 0x93, 0xe1, 0x6f, 0x5c, 0x4, 0xd6, 0x81, 0x60, 0xcd, 0x15, 0x43, 0x41, 0x8, 0x79, 0xc3, 0x3f, 0xcd, 0x6b, 0xc7, 0x82, 0xd7, 0x68, 0xc7, 0xfa, 0xd8, 0xae, 0xa1, 0x56, 0x69, 0xc9, 0x39, 0x90, 0x74, 0x5f, 0x37, 0xd6, 0xd0, 0xba, 0x11, 0xc6, 0x23, 0x61, 0x23, 0xd5, 0x16, 0xa, 0x4e, 0xb3, 0x2d, 0x5d, 0x67, 0xa7, 0x8f, 0x96, 0xb, 0x80, 0x22, 0x29, 0xac, 0x22, 0x1, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x2, 0x60, 0xfc, 0x3f, 0x2a, 0x93, 0xe1, 0x6f, 0x5c, 0x4, 0xd6, 0x81, 0x60, 0xcd, 0x15, 0x43, 0x41, 0x8, 0x79, 0xc3, 0x3f, 0xcd, 0x6b, 0xc7, 0x82, 0xd7, 0x68, 0xc7, 0xfa, 0xd8, 0xae, 0xa1, 0x56, 0x69, 0xc9, 0x39, 0x90, 0x74, 0x5f, 0x37, 0xd6, 0xd0, 0xba, 0x11, 0xc6, 0x23, 0x61, 0x23, 0xd5, 0x16, 0xa, 0x4e, 0xb3, 0x2d, 0x5d, 0x67, 0xa7, 0x8f, 0x96, 0xb, 0x80, 0x22, 0x29, 0xac, 0x22, 0x1, 0x3d, 0xf9, 0x59, 0x4f, 0x13, 0x8c, 0xf1, 0xb, 0x8, 0xbd, 0xb2, 0xca, 0x69, 0xe0, 0xbf, 0xd2, 0x58, 0x34, 0xc5, 0x5f, 0x47, 0x3c, 0xc, 0x91, 0x8d, 0x95, 0x59, 0x38, 0x80, 0x67, 0xa, 0x4a, 0x28, 0x5c, 0x10, 0xf7, 0x63, 0x7d, 0x89, 0x77, 0x3, 0xbe, 0xb3, 0x25, 0xc1, 0xa5, 0x84, 0x5d, 0xa8, 0x90, 0x87, 0x4d, 0x38, 0x1, 0x10, 0xc6, 0x80, 0xa1, 0xcf, 0xc1, 0x64, 0x40, 0xb7, 0xd6, 0x1},
		},
		{
			name: "AdvanceTimeTx",
			createSignedTx: func() (*signed.Tx, error) {
				return h.txBuilder.NewAdvanceTimeTx(time.Unix(286331153, 0))
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x13, 0x0, 0x0, 0x0, 0x0, 0x11, 0x11, 0x11, 0x11},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x13, 0x0, 0x0, 0x0, 0x0, 0x11, 0x11, 0x11, 0x11, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name: "CreateChain",
			createSignedTx: func() (*signed.Tx, error) {
				return h.txBuilder.NewCreateChainTx(
					testSubnet1.ID(),
					nil,
					constants.AVMID,
					nil,
					"chain name",
					[]*crypto.PrivateKeySECP256K1R{testSubnet1ControlKeys[0], testSubnet1ControlKeys[1]},
					ids.ShortEmpty,
				)
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f, 0x0, 0xa, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x20, 0x6e, 0x61, 0x6d, 0x65, 0x61, 0x76, 0x6d, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f, 0x0, 0xa, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x20, 0x6e, 0x61, 0x6d, 0x65, 0x61, 0x76, 0x6d, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x2, 0x5a, 0x6e, 0x9, 0x0, 0x4e, 0x80, 0x66, 0xeb, 0x73, 0xf7, 0xca, 0x6a, 0x35, 0x4b, 0x17, 0x9, 0x58, 0x37, 0xdd, 0xd7, 0xfc, 0x51, 0x4e, 0xfc, 0x48, 0x74, 0x92, 0x82, 0xd7, 0xb1, 0xca, 0x47, 0x51, 0x17, 0xcf, 0xe7, 0xf1, 0xfb, 0x85, 0xdd, 0x46, 0x8, 0x55, 0x61, 0x15, 0xc8, 0x11, 0x80, 0x96, 0x42, 0x6b, 0xaf, 0x6c, 0x29, 0x6d, 0xb8, 0x4c, 0x6c, 0x67, 0x23, 0x83, 0x39, 0xe1, 0x57, 0x0, 0xc9, 0xc, 0x53, 0x6a, 0x22, 0x82, 0x57, 0x36, 0xbf, 0xc4, 0x8a, 0x65, 0xc8, 0xd5, 0x3c, 0x1e, 0xd2, 0x6c, 0xc5, 0x97, 0x6f, 0x86, 0x2d, 0xdb, 0x36, 0xc2, 0x6e, 0xd4, 0x4f, 0xdc, 0xd1, 0xa4, 0x7c, 0x18, 0x5b, 0x0, 0x49, 0x91, 0xf3, 0x96, 0x2, 0xea, 0x32, 0xbc, 0xb0, 0x42, 0xea, 0xda, 0xb0, 0x4a, 0xf0, 0x59, 0x2d, 0xde, 0xac, 0x50, 0xd7, 0xf0, 0xde, 0x50, 0x33, 0x22, 0xb1, 0x31, 0x0},
		},
		{
			name: "CreateSubnet",
			createSignedTx: func() (*signed.Tx, error) {
				return h.txBuilder.NewCreateSubnetTx(
					2, // threshold; 2 sigs from keys[0], keys[1], keys[2] needed to add validator to this subnet
					[]ids.ShortID{ // control keys
						preFundedKeys[0].PublicKey().Address(),
						preFundedKeys[1].PublicKey().Address(),
						preFundedKeys[2].PublicKey().Address(),
					},
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
					preFundedKeys[0].PublicKey().Address(),
				)
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x3, 0x6e, 0xad, 0x69, 0x3c, 0x17, 0xab, 0xb1, 0xbe, 0x42, 0x2b, 0xb5, 0xb, 0x30, 0xb9, 0x71, 0x1f, 0xf9, 0x8d, 0x66, 0x7e, 0xf2, 0x42, 0x8, 0x46, 0x87, 0x6e, 0x69, 0xf4, 0x73, 0xdd, 0xa2, 0x56, 0x17, 0x29, 0x67, 0xe9, 0x92, 0xf0, 0xee, 0x31, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x3, 0x6e, 0xad, 0x69, 0x3c, 0x17, 0xab, 0xb1, 0xbe, 0x42, 0x2b, 0xb5, 0xb, 0x30, 0xb9, 0x71, 0x1f, 0xf9, 0x8d, 0x66, 0x7e, 0xf2, 0x42, 0x8, 0x46, 0x87, 0x6e, 0x69, 0xf4, 0x73, 0xdd, 0xa2, 0x56, 0x17, 0x29, 0x67, 0xe9, 0x92, 0xf0, 0xee, 0x31, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name: "Export",
			createSignedTx: func() (*signed.Tx, error) {
				return h.txBuilder.NewExportTx( // Test GetTx works for proposal blocks
					100,
					h.ctx.XChainID,
					preFundedKeys[0].PublicKey().Address(),
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
					preFundedKeys[0].PublicKey().Address(), // change addr
				)
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x64, 0x38, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x34, 0xce, 0x1d, 0xf2, 0x3b, 0x83, 0x8c, 0x5a, 0xbf, 0x2a, 0x7f, 0x64, 0x37, 0xcc, 0xa3, 0xd3, 0x6, 0x7e, 0xd5, 0x9, 0xff, 0x25, 0xf1, 0x1d, 0xf6, 0xb1, 0x1b, 0x58, 0x2b, 0x51, 0xeb, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x64, 0x38, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x1d, 0xcd, 0x65, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x34, 0xce, 0x1d, 0xf2, 0x3b, 0x83, 0x8c, 0x5a, 0xbf, 0x2a, 0x7f, 0x64, 0x37, 0xcc, 0xa3, 0xd3, 0x6, 0x7e, 0xd5, 0x9, 0xff, 0x25, 0xf1, 0x1d, 0xf6, 0xb1, 0x1b, 0x58, 0x2b, 0x51, 0xeb, 0x0, 0x0, 0x0, 0x1, 0x79, 0x65, 0x65, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x1, 0xc, 0xdf, 0xb9, 0x23, 0x2, 0xc6, 0x37, 0xcb, 0x7f, 0x39, 0xdd, 0xe2, 0xba, 0x22, 0x83, 0x3f, 0xa2, 0x4d, 0xc6, 0x34, 0xf8, 0x85, 0x57, 0x3c, 0x47, 0x32, 0x5c, 0xd, 0x82, 0xe5, 0x4c, 0xac, 0x10, 0x26, 0x4a, 0xbd, 0x75, 0x84, 0x72, 0x9d, 0xd5, 0x46, 0xea, 0x9a, 0x7a, 0x11, 0x47, 0x3b, 0x45, 0xf4, 0xd0, 0x68, 0xd1, 0x33, 0xff, 0xb1, 0xb7, 0x67, 0x8b, 0x96, 0xf3, 0xd, 0xf2, 0x8c, 0x1},
		},
		{
			name: "Import",
			createSignedTx: func() (*signed.Tx, error) {
				utx := &unsigned.ImportTx{
					BaseTx: unsigned.BaseTx{BaseTx: avax.BaseTx{
						NetworkID:    uint32(2022),
						BlockchainID: ids.ID{'B', 'l', 'o', 'c', 'k', 'c', 'h', 'a', 'i', 'n'},
						Outs: []*avax.TransferableOutput{{
							Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
							Out: &secp256k1fx.TransferOutput{
								Amt: uint64(1234),
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
								},
							},
						}},
						Ins: []*avax.TransferableInput{{
							UTXOID: avax.UTXOID{
								TxID:        ids.ID{'t', 'x', 'I', 'D'},
								OutputIndex: 2,
							},
							Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
							In: &secp256k1fx.TransferInput{
								Amt:   uint64(5678),
								Input: secp256k1fx.Input{SigIndices: []uint32{0}},
							},
						}},
					}},
					SourceChain: ids.ID{'S', 'o', 'u', 'r', 'c', 'e', 'c', 'h', 'a', 'i', 'n'},
				}
				tx := &signed.Tx{Unsigned: utx}
				signers := [][]*crypto.PrivateKeySECP256K1R(nil)
				if err := tx.Sign(unsigned.Codec, signers); err != nil {
					return nil, err
				}
				return tx, nil
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x0, 0x0, 0x7, 0xe6, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4, 0xd2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x1, 0x74, 0x78, 0x49, 0x44, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x16, 0x2e, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x0, 0x0, 0x7, 0xe6, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4, 0xd2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0xfc, 0xed, 0xa8, 0xf9, 0xf, 0xcb, 0x5d, 0x30, 0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2, 0x93, 0x7, 0x76, 0x26, 0x0, 0x0, 0x0, 0x1, 0x74, 0x78, 0x49, 0x44, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x16, 0x2e, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name: "RewardValidator",
			createSignedTx: func() (*signed.Tx, error) {
				return h.txBuilder.NewRewardValidatorTx(testSubnet1.ID())
			},
			unsignedBytes: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f},
			signedBytes:   []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x27, 0x11, 0xf8, 0xc0, 0xcd, 0xdd, 0x6e, 0x16, 0xbe, 0xb8, 0xdd, 0xc4, 0x8b, 0xcd, 0xed, 0xf5, 0x6f, 0x4f, 0xee, 0xf9, 0xb2, 0xce, 0x43, 0xc2, 0xb1, 0x2e, 0x4d, 0xd9, 0x92, 0x90, 0xa0, 0x5f, 0x0, 0x0, 0x0, 0x0},
		},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		signedTx, err := tt.createSignedTx()
		assert.NoError(err)
		assert.Equal(tt.unsignedBytes, signedTx.Unsigned.UnsignedBytes())
		assert.Equal(tt.signedBytes, signedTx.Bytes())
	}
}
