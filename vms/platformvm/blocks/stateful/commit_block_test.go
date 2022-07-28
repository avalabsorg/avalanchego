// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateful/version"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/stretchr/testify/assert"
)

func testProposalTx() (*txs.Tx, error) {
	utx := &txs.RewardValidatorTx{
		TxID: ids.ID{'r', 'e', 'w', 'a', 'r', 'd', 'I', 'D'},
	}

	signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
	return txs.NewSigned(utx, txs.Codec, signers)
}

func TestBlueberryCommitBlockTimestampChecks(t *testing.T) {
	assert := assert.New(t)

	h := newTestHelpersCollection(t, nil)
	defer func() {
		if err := internalStateShutdown(h); err != nil {
			t.Fatal(err)
		}
	}()

	now := defaultGenesisTime.Add(time.Hour)
	h.clk.Set(now)
	blkVersion := uint16(version.BlueberryBlockVersion)

	tests := []struct {
		description string
		parentTime  time.Time
		childTime   time.Time
		result      error
	}{
		{
			description: "commit block timestamp matching parent's one",
			parentTime:  now,
			childTime:   now,
			result:      nil,
		},
		{
			description: "commit block timestamp before parent's one",
			childTime:   now.Add(-1 * time.Second),
			parentTime:  now,
			result:      errOptionBlockTimestampNotMatchingParent,
		},
		{
			description: "commit block timestamp after parent's one",
			parentTime:  now,
			childTime:   now.Add(time.Second),
			result:      errOptionBlockTimestampNotMatchingParent,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h.cfg.BlueberryTime = time.Time{} // activate Blueberry

			// setup and store parent block
			// it's a standard block for simplicity
			parentVersion := blkVersion
			parentHeight := uint64(2022)
			parentTx, err := testProposalTx()
			assert.NoError(err)

			blueberryParentBlk, err := stateless.NewProposalBlock(
				parentVersion,
				uint64(test.parentTime.Unix()),
				ids.Empty, // parentID does not matter
				parentHeight,
				parentTx,
			)
			assert.NoError(err)
			h.fullState.AddStatelessBlock(blueberryParentBlk, choices.Accepted)

			// build and verify child block
			childVersion := blkVersion
			childHeight := parentHeight + 1
			statelessCommitBlk, err := stateless.NewCommitBlock(
				childVersion,
				uint64(test.childTime.Unix()),
				blueberryParentBlk.ID(),
				childHeight,
			)
			assert.NoError(err)

			// call verify on it
			commonBlk := statelessCommitBlk.(*stateless.CommitBlock)
			err = h.blkManager.(*manager).verifier.(*verifier).verifyCommonBlock(commonBlk)
			assert.ErrorIs(err, test.result)
		})
	}
}
