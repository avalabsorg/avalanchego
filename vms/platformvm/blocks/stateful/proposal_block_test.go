// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateful

import (
	"fmt"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateful/version"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/reward"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/executor"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestApricotProposalBlockTimeVerification(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	env := newEnvironment(t, ctrl)
	defer func() {
		if err := shutdownEnvironment(env); err != nil {
			t.Fatal(err)
		}
	}()

	// create apricotParentBlk. It's a standard one for simplicity
	parentUnixTime := uint64(0)
	parentHeight := uint64(2022)

	apricotParentBlk, err := stateless.NewStandardBlock(
		stateless.ApricotVersion,
		parentUnixTime,
		ids.Empty, // does not matter
		parentHeight,
		nil, // txs do not matter in this test
	)
	assert.NoError(err)
	parentID := apricotParentBlk.ID()

	// store parent block, with relevant quantities
	env.blkManager.(*manager).blkIDToState[parentID] = &blockState{
		statelessBlock: apricotParentBlk,
	}
	env.blkManager.(*manager).lastAccepted = parentID
	env.blkManager.(*manager).stateVersions.SetState(parentID, env.mockedState)
	env.mockedState.EXPECT().GetLastAccepted().Return(parentID).AnyTimes()

	// create a proposal transaction to be included into proposal block
	chainTime := env.clk.Time().Truncate(time.Second)
	utx := &txs.AddValidatorTx{
		BaseTx:       txs.BaseTx{},
		Validator:    validator.Validator{End: uint64(chainTime.Unix())},
		Stake:        nil,
		RewardsOwner: &secp256k1fx.OutputOwners{},
		Shares:       uint32(defaultTxFee),
	}
	addValTx := &txs.Tx{Unsigned: utx}
	assert.NoError(addValTx.Sign(txs.Codec, nil))
	blkTx := &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addValTx.ID(),
		},
	}

	// setup state to validate proposal block transaction
	env.mockedState.EXPECT().GetTimestamp().Return(chainTime).AnyTimes()

	currentStakersIt := state.NewMockStakerIterator(ctrl)
	currentStakersIt.EXPECT().Next().Return(true)
	currentStakersIt.EXPECT().Value().Return(&state.Staker{
		TxID:    addValTx.ID(),
		EndTime: chainTime,
	})
	currentStakersIt.EXPECT().Release()
	env.mockedState.EXPECT().GetCurrentStakerIterator().Return(currentStakersIt, nil)
	env.mockedState.EXPECT().GetTx(addValTx.ID()).Return(addValTx, status.Committed, nil)

	env.mockedState.EXPECT().GetCurrentSupply().Return(uint64(1000)).AnyTimes()
	env.mockedState.EXPECT().GetUptime(gomock.Any()).Return(
		time.Duration(1000), /*upDuration*/
		time.Time{},         /*lastUpdated*/
		nil,                 /*err*/
	).AnyTimes()

	// wrong height
	statelessProposalBlock, err := stateless.NewProposalBlock(
		stateless.ApricotVersion,
		parentUnixTime,
		parentID,
		parentHeight,
		blkTx,
	)
	block := env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// valid
	statelessProposalBlock, err = stateless.NewProposalBlock(
		stateless.ApricotVersion,
		parentUnixTime,
		parentID,
		parentHeight+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.NoError(block.Verify())
}

func TestBlueberryProposalBlockTimeVerification(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	env := newEnvironment(t, ctrl)
	defer func() {
		if err := shutdownEnvironment(env); err != nil {
			t.Fatal(err)
		}
	}()
	env.clk.Set(defaultGenesisTime)
	env.config.BlueberryTime = time.Time{} // activate Blueberry

	// create parentBlock. It's a standard one for simplicity
	blksVersion := uint16(stateless.ApricotVersion)
	parentTime := defaultGenesisTime
	parentHeight := uint64(2022)

	blueberryParentBlk, err := stateless.NewStandardBlock(
		blksVersion,
		uint64(parentTime.Unix()),
		genesisBlkID, // does not matter
		parentHeight,
		nil, // txs do not matter in this test
	)
	assert.NoError(err)
	parentID := blueberryParentBlk.ID()

	// store parent block, with relevant quantities
	chainTime := parentTime
	env.mockedState.EXPECT().GetTimestamp().Return(chainTime).AnyTimes()
	env.mockedState.EXPECT().GetCurrentSupply().Return(uint64(1000)).AnyTimes()

	onAcceptState, err := state.NewDiff(genesisBlkID, env.backend.StateVersions)
	assert.NoError(err)
	onAcceptState.SetTimestamp(parentTime)

	env.blkManager.(*manager).blkIDToState[parentID] = &blockState{
		statelessBlock: blueberryParentBlk,
		onAcceptState:  onAcceptState,
		timestamp:      parentTime,
	}
	env.blkManager.(*manager).lastAccepted = parentID
	env.blkManager.(*manager).stateVersions.SetState(parentID, env.mockedState)
	env.mockedState.EXPECT().GetLastAccepted().Return(parentID).AnyTimes()
	env.mockedState.EXPECT().GetStatelessBlock(gomock.Any()).DoAndReturn(
		func(blockID ids.ID) (stateless.Block, choices.Status, error) {
			if blockID == parentID {
				return blueberryParentBlk, choices.Accepted, nil
			}
			return nil, choices.Rejected, database.ErrNotFound
		}).AnyTimes()

	// setup state to validate proposal block transaction
	nextStakerTime := chainTime.Add(executor.SyncBound).Add(-1 * time.Second)
	nextStakerTx := &txs.Tx{
		Unsigned: &txs.AddValidatorTx{
			BaseTx:       txs.BaseTx{},
			Validator:    validator.Validator{End: uint64(nextStakerTime.Unix())},
			Stake:        nil,
			RewardsOwner: &secp256k1fx.OutputOwners{},
			Shares:       uint32(defaultTxFee),
		},
	}
	assert.NoError(nextStakerTx.Sign(txs.Codec, nil))
	nextStakerTxID := nextStakerTx.ID()
	env.mockedState.EXPECT().GetTx(nextStakerTxID).Return(nextStakerTx, status.Processing, nil)

	currentStakersIt := state.NewMockStakerIterator(ctrl)
	currentStakersIt.EXPECT().Next().Return(true).AnyTimes()
	currentStakersIt.EXPECT().Value().Return(&state.Staker{
		TxID:     nextStakerTxID,
		EndTime:  nextStakerTime,
		NextTime: nextStakerTime,
		Priority: state.PrimaryNetworkValidatorCurrentPriority,
	}).AnyTimes()
	currentStakersIt.EXPECT().Release().AnyTimes()
	env.mockedState.EXPECT().GetCurrentStakerIterator().Return(currentStakersIt, nil).AnyTimes()

	pendingStakersIt := state.NewMockStakerIterator(ctrl)
	pendingStakersIt.EXPECT().Next().Return(false).AnyTimes() // no pending stakers
	pendingStakersIt.EXPECT().Release().AnyTimes()
	env.mockedState.EXPECT().GetPendingStakerIterator().Return(pendingStakersIt, nil).AnyTimes()

	env.mockedState.EXPECT().GetUptime(gomock.Any()).Return(
		time.Duration(1000), /*upDuration*/
		time.Time{},         /*lastUpdated*/
		nil,                 /*err*/
	).AnyTimes()

	// create proposal tx to be included in the proposal block
	blkTx := &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: nextStakerTxID,
		},
	}
	assert.NoError(blkTx.Sign(txs.Codec, nil))

	// wrong height
	statelessProposalBlock, err := stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(parentTime.Add(time.Second).Unix()),
		parentID,
		blueberryParentBlk.Height(),
		blkTx,
	)
	block := env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// wrong version
	statelessProposalBlock, err = stateless.NewProposalBlock(
		stateless.ApricotVersion,
		uint64(parentTime.Add(time.Second).Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// wrong timestamp, earlier than parent
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(parentTime.Add(-1*time.Second).Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// wrong timestamp, violated synchrony bound
	beyondSyncBoundTimeStamp := env.clk.Time().Add(executor.SyncBound).Add(time.Second)
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(beyondSyncBoundTimeStamp.Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// wrong timestamp, skipped staker set change event
	skippedStakerEventTimeStamp := nextStakerTime.Add(time.Second)
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(skippedStakerEventTimeStamp.Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// wrong tx content (no advance time txs)
	invalidTx := &txs.Tx{
		Unsigned: &txs.AdvanceTimeTx{
			Time: uint64(nextStakerTime.Unix()),
		},
	}
	assert.NoError(invalidTx.Sign(txs.Codec, nil))
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(parentTime.Add(time.Second).Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		invalidTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.Error(block.Verify())

	// valid
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(nextStakerTime.Unix()),
		parentID,
		blueberryParentBlk.Height()+1,
		blkTx,
	)
	block = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.NoError(block.Verify())
}

func TestBlueberryProposalBlockUpdateStakers(t *testing.T) {
	// Chronological order (not in scale):
	// Staker0:    |--- ??? // Staker0 end time depends on the test
	// Staker1:        |------------------------------------------------------------------------|
	// Staker2:            |------------------------|
	// Staker3:                |------------------------|
	// Staker3sub:                 |----------------|
	// Staker4:                |------------------------|
	// Staker5:                                     |------------------------|

	// Staker0 it's here just to allow to issue a proposal block with the chosen endTime.
	staker0RewardAddress := ids.GenerateTestShortID()
	staker0 := staker{
		nodeID:        ids.NodeID(staker0RewardAddress),
		rewardAddress: staker0RewardAddress,
		startTime:     defaultGenesisTime,
		endTime:       time.Time{}, // actual endTime depends on specific test
	}

	staker1RewardAddress := ids.GenerateTestShortID()
	staker1 := staker{
		nodeID:        ids.NodeID(staker1RewardAddress),
		rewardAddress: staker1RewardAddress,
		startTime:     defaultGenesisTime.Add(1 * time.Minute),
		endTime:       defaultGenesisTime.Add(10 * defaultMinStakingDuration).Add(1 * time.Minute),
	}

	staker2RewardAddress := ids.GenerateTestShortID()
	staker2 := staker{
		nodeID:        ids.NodeID(staker2RewardAddress),
		rewardAddress: staker2RewardAddress,
		startTime:     staker1.startTime.Add(1 * time.Minute),
		endTime:       staker1.startTime.Add(1 * time.Minute).Add(defaultMinStakingDuration),
	}

	staker3RewardAddress := ids.GenerateTestShortID()
	staker3 := staker{
		nodeID:        ids.NodeID(staker3RewardAddress),
		rewardAddress: staker3RewardAddress,
		startTime:     staker2.startTime.Add(1 * time.Minute),
		endTime:       staker2.endTime.Add(1 * time.Minute),
	}

	staker3Sub := staker{
		nodeID:        staker3.nodeID,
		rewardAddress: staker3.rewardAddress,
		startTime:     staker3.startTime.Add(1 * time.Minute),
		endTime:       staker3.endTime.Add(-1 * time.Minute),
	}

	staker4RewardAddress := ids.GenerateTestShortID()
	staker4 := staker{
		nodeID:        ids.NodeID(staker4RewardAddress),
		rewardAddress: staker4RewardAddress,
		startTime:     staker3.startTime,
		endTime:       staker3.endTime,
	}

	staker5RewardAddress := ids.GenerateTestShortID()
	staker5 := staker{
		nodeID:        ids.NodeID(staker5RewardAddress),
		rewardAddress: staker5RewardAddress,
		startTime:     staker2.endTime,
		endTime:       staker2.endTime.Add(defaultMinStakingDuration),
	}

	tests := []test{
		{
			description:   "advance time to before staker1 start with subnet",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			subnetStakers: []staker{staker1, staker2, staker3, staker4, staker5},
			advanceTimeTo: []time.Time{staker1.startTime.Add(-1 * time.Second)},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: pending,
				staker2.nodeID: pending,
				staker3.nodeID: pending,
				staker4.nodeID: pending,
				staker5.nodeID: pending,
			},
			expectedSubnetStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: pending,
				staker2.nodeID: pending,
				staker3.nodeID: pending,
				staker4.nodeID: pending,
				staker5.nodeID: pending,
			},
		},
		{
			description:   "advance time to staker 1 start with subnet",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			subnetStakers: []staker{staker1},
			advanceTimeTo: []time.Time{staker1.startTime},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: pending,
				staker3.nodeID: pending,
				staker4.nodeID: pending,
				staker5.nodeID: pending,
			},
			expectedSubnetStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: pending,
				staker3.nodeID: pending,
				staker4.nodeID: pending,
				staker5.nodeID: pending,
			},
		},
		{
			description:   "advance time to the staker2 start",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			advanceTimeTo: []time.Time{staker1.startTime, staker2.startTime},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: current,
				staker3.nodeID: pending,
				staker4.nodeID: pending,
				staker5.nodeID: pending,
			},
		},
		{
			description:   "staker3 should validate only primary network",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			subnetStakers: []staker{staker1, staker2, staker3Sub, staker4, staker5},
			advanceTimeTo: []time.Time{staker1.startTime, staker2.startTime, staker3.startTime},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: current,
				staker3.nodeID: current,
				staker4.nodeID: current,
				staker5.nodeID: pending,
			},
			expectedSubnetStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID:    current,
				staker2.nodeID:    current,
				staker3Sub.nodeID: pending,
				staker4.nodeID:    current,
				staker5.nodeID:    pending,
			},
		},
		{
			description:   "advance time to staker3 start with subnet",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			subnetStakers: []staker{staker1, staker2, staker3Sub, staker4, staker5},
			advanceTimeTo: []time.Time{staker1.startTime, staker2.startTime, staker3.startTime, staker3Sub.startTime},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: current,
				staker3.nodeID: current,
				staker4.nodeID: current,
				staker5.nodeID: pending,
			},
			expectedSubnetStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,
				staker2.nodeID: current,
				staker3.nodeID: current,
				staker4.nodeID: current,
				staker5.nodeID: pending,
			},
		},
		{
			description:   "advance time to staker5 end",
			stakers:       []staker{staker1, staker2, staker3, staker4, staker5},
			advanceTimeTo: []time.Time{staker1.startTime, staker2.startTime, staker3.startTime, staker5.startTime},
			expectedStakers: map[ids.NodeID]stakerStatus{
				staker1.nodeID: current,

				// given its txID, staker2 will be
				// rewarded and moved out of current stakers set
				// staker2.nodeID: current,
				staker3.nodeID: current,
				staker4.nodeID: current,
				staker5.nodeID: current,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(ts *testing.T) {
			assert := assert.New(ts)
			env := newEnvironment(t, nil)
			defer func() {
				if err := shutdownEnvironment(env); err != nil {
					t.Fatal(err)
				}
			}()

			env.config.BlueberryTime = time.Time{} // activate Blueberry
			env.config.WhitelistedSubnets.Add(testSubnet1.ID())

			for _, staker := range test.stakers {
				tx, err := env.txBuilder.NewAddValidatorTx(
					env.config.MinValidatorStake,
					uint64(staker.startTime.Unix()),
					uint64(staker.endTime.Unix()),
					staker.nodeID,
					staker.rewardAddress,
					reward.PercentDenominator,
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
					ids.ShortEmpty,
				)
				assert.NoError(err)

				staker := state.NewPrimaryNetworkStaker(tx.ID(), &tx.Unsigned.(*txs.AddValidatorTx).Validator)
				staker.NextTime = staker.StartTime
				staker.Priority = state.SubnetValidatorPendingPriority

				env.state.PutPendingValidator(staker)
				env.state.AddTx(tx, status.Committed)
				assert.NoError(env.state.Commit())
			}

			for _, subStaker := range test.subnetStakers {
				tx, err := env.txBuilder.NewAddSubnetValidatorTx(
					10, // Weight
					uint64(subStaker.startTime.Unix()),
					uint64(subStaker.endTime.Unix()),
					subStaker.nodeID, // validator ID
					testSubnet1.ID(), // Subnet ID
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
					ids.ShortEmpty,
				)
				assert.NoError(err)

				subnetStaker := state.NewSubnetStaker(tx.ID(), &tx.Unsigned.(*txs.AddSubnetValidatorTx).Validator)
				subnetStaker.NextTime = subStaker.startTime
				subnetStaker.Priority = state.SubnetValidatorPendingPriority

				env.state.PutPendingValidator(subnetStaker)
				env.state.AddTx(tx, status.Committed)
				assert.NoError(env.state.Commit())
			}

			for _, newTime := range test.advanceTimeTo {
				env.clk.Set(newTime)

				// add Staker0 (with the right end time) to state
				// so to allow proposalBlk issuance
				staker0.endTime = newTime
				addStaker0, err := env.txBuilder.NewAddValidatorTx(
					10,
					uint64(staker0.startTime.Unix()),
					uint64(staker0.endTime.Unix()),
					staker0.nodeID,
					staker0.rewardAddress,
					reward.PercentDenominator,
					[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
					ids.ShortEmpty,
				)
				assert.NoError(err)

				// store Staker0 to state
				staker0 := state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
				staker0.NextTime = staker0.EndTime
				staker0.Priority = state.PrimaryNetworkValidatorCurrentPriority
				env.state.PutCurrentValidator(staker0)
				env.state.AddTx(addStaker0, status.Committed)
				assert.NoError(env.state.Commit())

				s0RewardTx := &txs.Tx{
					Unsigned: &txs.RewardValidatorTx{
						TxID: staker0.TxID,
					},
				}
				assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

				// build proposal block moving ahead chain time
				// as well as rewarding staker0
				preferredID := env.state.GetLastAccepted()
				parentBlk, _, err := env.state.GetStatelessBlock(preferredID)
				assert.NoError(err)
				statelessProposalBlock, err := stateless.NewProposalBlock(
					version.BlueberryBlockVersion,
					uint64(newTime.Unix()),
					parentBlk.ID(),
					parentBlk.Height()+1,
					s0RewardTx,
				)
				assert.NoError(err)

				// verify and accept the block
				block := env.blkManager.NewBlock(statelessProposalBlock)
				assert.NoError(block.Verify())
				options, err := block.(snowman.OracleBlock).Options()
				assert.NoError(err)

				assert.NoError(options[0].Verify())

				assert.NoError(block.Accept())
				assert.NoError(options[0].Accept())
			}
			assert.NoError(env.state.Commit())

			for stakerNodeID, status := range test.expectedStakers {
				switch status {
				case pending:
					_, err := env.state.GetPendingValidator(constants.PrimaryNetworkID, stakerNodeID)
					assert.NoError(err)
					assert.False(env.config.Validators.Contains(constants.PrimaryNetworkID, stakerNodeID))
				case current:
					_, err := env.state.GetCurrentValidator(constants.PrimaryNetworkID, stakerNodeID)
					assert.NoError(err)
					assert.True(env.config.Validators.Contains(constants.PrimaryNetworkID, stakerNodeID))
				}
			}

			for stakerNodeID, status := range test.expectedSubnetStakers {
				switch status {
				case pending:
					assert.False(env.config.Validators.Contains(testSubnet1.ID(), stakerNodeID))
				case current:
					assert.True(env.config.Validators.Contains(testSubnet1.ID(), stakerNodeID))
				}
			}
		})
	}
}

func TestBlueberryProposalBlockRemoveSubnetValidator(t *testing.T) {
	assert := assert.New(t)
	env := newEnvironment(t, nil)
	defer func() {
		if err := shutdownEnvironment(env); err != nil {
			t.Fatal(err)
		}
	}()
	env.config.BlueberryTime = time.Time{} // activate Blueberry
	env.config.WhitelistedSubnets.Add(testSubnet1.ID())

	// Add a subnet validator to the staker set
	subnetValidatorNodeID := ids.NodeID(preFundedKeys[0].PublicKey().Address())
	// Starts after the corre
	subnetVdr1StartTime := defaultValidateStartTime
	subnetVdr1EndTime := defaultValidateStartTime.Add(defaultMinStakingDuration)
	tx, err := env.txBuilder.NewAddSubnetValidatorTx(
		1,                                  // Weight
		uint64(subnetVdr1StartTime.Unix()), // Start time
		uint64(subnetVdr1EndTime.Unix()),   // end time
		subnetValidatorNodeID,              // Node ID
		testSubnet1.ID(),                   // Subnet ID
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	staker := state.NewSubnetStaker(tx.ID(), &tx.Unsigned.(*txs.AddSubnetValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.SubnetValidatorCurrentPriority

	env.state.PutCurrentValidator(staker)
	env.state.AddTx(tx, status.Committed)
	assert.NoError(env.state.Commit())

	// The above validator is now part of the staking set

	// Queue a staker that joins the staker set after the above validator leaves
	subnetVdr2NodeID := ids.NodeID(preFundedKeys[1].PublicKey().Address())
	tx, err = env.txBuilder.NewAddSubnetValidatorTx(
		1, // Weight
		uint64(subnetVdr1EndTime.Add(time.Second).Unix()),                                // Start time
		uint64(subnetVdr1EndTime.Add(time.Second).Add(defaultMinStakingDuration).Unix()), // end time
		subnetVdr2NodeID, // Node ID
		testSubnet1.ID(), // Subnet ID
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	staker = state.NewSubnetStaker(tx.ID(), &tx.Unsigned.(*txs.AddSubnetValidatorTx).Validator)
	staker.NextTime = staker.StartTime
	staker.Priority = state.SubnetValidatorPendingPriority

	env.state.PutPendingValidator(staker)
	env.state.AddTx(tx, status.Committed)
	assert.NoError(env.state.Commit())

	// The above validator is now in the pending staker set

	// Advance time to the first staker's end time.
	env.clk.Set(subnetVdr1EndTime)

	// add Staker0 (with the right end time) to state
	// so to allow proposalBlk issuance
	staker0StartTime := defaultValidateStartTime
	staker0EndTime := subnetVdr1EndTime
	addStaker0, err := env.txBuilder.NewAddValidatorTx(
		10,
		uint64(staker0StartTime.Unix()),
		uint64(staker0EndTime.Unix()),
		ids.GenerateTestNodeID(),
		ids.GenerateTestShortID(),
		reward.PercentDenominator,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	// store Staker0 to state
	staker = state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
	env.state.PutCurrentValidator(staker)
	env.state.AddTx(addStaker0, status.Committed)
	assert.NoError(env.state.Commit())

	// create rewardTx for staker0
	s0RewardTx := &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addStaker0.ID(),
		},
	}
	assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

	// build proposal block moving ahead chain time
	preferredID := env.state.GetLastAccepted()
	parentBlk, _, err := env.state.GetStatelessBlock(preferredID)
	assert.NoError(err)
	statelessProposalBlock, err := stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(subnetVdr1EndTime.Unix()),
		parentBlk.ID(),
		parentBlk.Height()+1,
		s0RewardTx,
	)
	assert.NoError(err)
	propBlk := env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(propBlk.Verify()) // verify and update staker set

	options, err := propBlk.(snowman.OracleBlock).Options()
	assert.NoError(err)
	commitBlk := options[0]
	assert.NoError(commitBlk.Verify())

	blkStateMap := env.blkManager.(*manager).blkIDToState
	updatedState := blkStateMap[commitBlk.ID()].onAcceptState
	_, err = updatedState.GetCurrentValidator(testSubnet1.ID(), subnetValidatorNodeID)
	assert.ErrorIs(err, database.ErrNotFound)

	// Check VM Validators are removed successfully
	assert.NoError(propBlk.Accept())
	assert.NoError(commitBlk.Accept())
	assert.False(env.config.Validators.Contains(testSubnet1.ID(), subnetVdr2NodeID))
	assert.False(env.config.Validators.Contains(testSubnet1.ID(), subnetValidatorNodeID))
}

func TestBlueberryProposalBlockWhitelistedSubnet(t *testing.T) {
	assert := assert.New(t)

	for _, whitelist := range []bool{true, false} {
		t.Run(fmt.Sprintf("whitelisted %t", whitelist), func(ts *testing.T) {
			env := newEnvironment(t, nil)
			defer func() {
				if err := shutdownEnvironment(env); err != nil {
					t.Fatal(err)
				}
			}()
			env.config.BlueberryTime = time.Time{} // activate Blueberry
			if whitelist {
				env.config.WhitelistedSubnets.Add(testSubnet1.ID())
			}

			// Add a subnet validator to the staker set
			subnetValidatorNodeID := ids.NodeID(preFundedKeys[0].PublicKey().Address())

			subnetVdr1StartTime := defaultGenesisTime.Add(1 * time.Minute)
			subnetVdr1EndTime := defaultGenesisTime.Add(10 * defaultMinStakingDuration).Add(1 * time.Minute)
			tx, err := env.txBuilder.NewAddSubnetValidatorTx(
				1,                                  // Weight
				uint64(subnetVdr1StartTime.Unix()), // Start time
				uint64(subnetVdr1EndTime.Unix()),   // end time
				subnetValidatorNodeID,              // Node ID
				testSubnet1.ID(),                   // Subnet ID
				[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
				ids.ShortEmpty,
			)
			assert.NoError(err)

			staker := state.NewSubnetStaker(tx.ID(), &tx.Unsigned.(*txs.AddSubnetValidatorTx).Validator)
			staker.NextTime = staker.StartTime
			staker.Priority = state.SubnetValidatorPendingPriority

			env.state.PutPendingValidator(staker)
			env.state.AddTx(tx, status.Committed)
			assert.NoError(env.state.Commit())

			// Advance time to the staker's start time.
			env.clk.Set(subnetVdr1StartTime)

			// add Staker0 (with the right end time) to state
			// so to allow proposalBlk issuance
			staker0StartTime := defaultGenesisTime
			staker0EndTime := subnetVdr1StartTime
			addStaker0, err := env.txBuilder.NewAddValidatorTx(
				10,
				uint64(staker0StartTime.Unix()),
				uint64(staker0EndTime.Unix()),
				ids.GenerateTestNodeID(),
				ids.GenerateTestShortID(),
				reward.PercentDenominator,
				[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
				ids.ShortEmpty,
			)
			assert.NoError(err)

			// store Staker0 to state
			staker = state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
			staker.NextTime = staker.EndTime
			staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
			env.state.PutCurrentValidator(staker)
			env.state.AddTx(addStaker0, status.Committed)
			assert.NoError(env.state.Commit())

			// create rewardTx for staker0
			s0RewardTx := &txs.Tx{
				Unsigned: &txs.RewardValidatorTx{
					TxID: addStaker0.ID(),
				},
			}
			assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

			// build proposal block moving ahead chain time
			preferredID := env.state.GetLastAccepted()
			parentBlk, _, err := env.state.GetStatelessBlock(preferredID)
			assert.NoError(err)
			statelessProposalBlock, err := stateless.NewProposalBlock(
				version.BlueberryBlockVersion,
				uint64(subnetVdr1StartTime.Unix()),
				parentBlk.ID(),
				parentBlk.Height()+1,
				s0RewardTx,
			)
			assert.NoError(err)
			propBlk := env.blkManager.NewBlock(statelessProposalBlock)
			assert.NoError(propBlk.Verify()) // verify update staker set
			options, err := propBlk.(snowman.OracleBlock).Options()
			assert.NoError(err)
			commitBlk := options[0]
			assert.NoError(commitBlk.Verify())

			assert.NoError(propBlk.Accept())
			assert.NoError(commitBlk.Accept())
			assert.Equal(whitelist, env.config.Validators.Contains(testSubnet1.ID(), subnetValidatorNodeID))
		})
	}
}

func TestBlueberryProposalBlockDelegatorStakerWeight(t *testing.T) {
	assert := assert.New(t)
	env := newEnvironment(t, nil)
	defer func() {
		if err := shutdownEnvironment(env); err != nil {
			t.Fatal(err)
		}
	}()
	env.config.BlueberryTime = time.Time{} // activate Blueberry

	// Case: Timestamp is after next validator start time
	// Add a pending validator
	pendingValidatorStartTime := defaultGenesisTime.Add(1 * time.Second)
	pendingValidatorEndTime := pendingValidatorStartTime.Add(defaultMaxStakingDuration)
	nodeID := ids.GenerateTestNodeID()
	rewardAddress := ids.GenerateTestShortID()
	_, err := addPendingValidator(
		env,
		pendingValidatorStartTime,
		pendingValidatorEndTime,
		nodeID,
		rewardAddress,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
	)
	assert.NoError(err)

	// add Staker0 (with the right end time) to state
	// just to allow proposalBlk issuance (with a reward Tx)
	staker0StartTime := defaultGenesisTime
	staker0EndTime := pendingValidatorStartTime
	addStaker0, err := env.txBuilder.NewAddValidatorTx(
		10,
		uint64(staker0StartTime.Unix()),
		uint64(staker0EndTime.Unix()),
		ids.GenerateTestNodeID(),
		ids.GenerateTestShortID(),
		reward.PercentDenominator,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	// store Staker0 to state
	staker := state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
	env.state.PutCurrentValidator(staker)
	env.state.AddTx(addStaker0, status.Committed)
	assert.NoError(env.state.Commit())

	// create rewardTx for staker0
	s0RewardTx := &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addStaker0.ID(),
		},
	}
	assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

	// build proposal block moving ahead chain time
	preferredID := env.state.GetLastAccepted()
	parentBlk, _, err := env.state.GetStatelessBlock(preferredID)
	assert.NoError(err)
	statelessProposalBlock, err := stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(pendingValidatorStartTime.Unix()),
		parentBlk.ID(),
		parentBlk.Height()+1,
		s0RewardTx,
	)
	assert.NoError(err)
	propBlk := env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(propBlk.Verify())

	options, err := propBlk.(snowman.OracleBlock).Options()
	assert.NoError(err)
	commitBlk := options[0]
	assert.NoError(commitBlk.Verify())

	assert.NoError(propBlk.Accept())
	assert.NoError(commitBlk.Accept())

	// Test validator weight before delegation
	primarySet, ok := env.config.Validators.GetValidators(constants.PrimaryNetworkID)
	assert.True(ok)
	vdrWeight, _ := primarySet.GetWeight(nodeID)
	assert.Equal(env.config.MinValidatorStake, vdrWeight)

	// Add delegator
	pendingDelegatorStartTime := pendingValidatorStartTime.Add(1 * time.Second)
	pendingDelegatorEndTime := pendingDelegatorStartTime.Add(1 * time.Second)

	addDelegatorTx, err := env.txBuilder.NewAddDelegatorTx(
		env.config.MinDelegatorStake,
		uint64(pendingDelegatorStartTime.Unix()),
		uint64(pendingDelegatorEndTime.Unix()),
		nodeID,
		preFundedKeys[0].PublicKey().Address(),
		[]*crypto.PrivateKeySECP256K1R{
			preFundedKeys[0],
			preFundedKeys[1],
			preFundedKeys[4],
		},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	staker = state.NewPrimaryNetworkStaker(addDelegatorTx.ID(), &addDelegatorTx.Unsigned.(*txs.AddDelegatorTx).Validator)
	staker.NextTime = staker.StartTime
	staker.Priority = state.PrimaryNetworkDelegatorPendingPriority

	env.state.PutPendingDelegator(staker)
	env.state.AddTx(addDelegatorTx, status.Committed)
	env.state.SetHeight( /*dummyHeight*/ uint64(1))
	assert.NoError(env.state.Commit())

	// add Staker0 (with the right end time) to state
	// so to allow proposalBlk issuance
	staker0EndTime = pendingDelegatorStartTime
	addStaker0, err = env.txBuilder.NewAddValidatorTx(
		10,
		uint64(staker0StartTime.Unix()),
		uint64(staker0EndTime.Unix()),
		ids.GenerateTestNodeID(),
		ids.GenerateTestShortID(),
		reward.PercentDenominator,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	// store Staker0 to state
	staker = state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
	env.state.PutCurrentValidator(staker)
	env.state.AddTx(addStaker0, status.Committed)
	assert.NoError(env.state.Commit())

	// create rewardTx for staker0
	s0RewardTx = &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addStaker0.ID(),
		},
	}
	assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

	// Advance Time
	preferredID = env.state.GetLastAccepted()
	parentBlk, _, err = env.state.GetStatelessBlock(preferredID)
	assert.NoError(err)
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(pendingDelegatorStartTime.Unix()),
		parentBlk.ID(),
		parentBlk.Height()+1,
		s0RewardTx,
	)
	propBlk = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(err)
	assert.NoError(propBlk.Verify())

	options, err = propBlk.(snowman.OracleBlock).Options()
	assert.NoError(err)
	commitBlk = options[0]
	assert.NoError(commitBlk.Verify())

	assert.NoError(propBlk.Accept())
	assert.NoError(commitBlk.Accept())

	// Test validator weight after delegation
	vdrWeight, _ = primarySet.GetWeight(nodeID)
	assert.Equal(env.config.MinDelegatorStake+env.config.MinValidatorStake, vdrWeight)
}

func TestBlueberryProposalBlockDelegatorStakers(t *testing.T) {
	assert := assert.New(t)
	env := newEnvironment(t, nil)
	defer func() {
		if err := shutdownEnvironment(env); err != nil {
			t.Fatal(err)
		}
	}()
	env.config.BlueberryTime = time.Time{} // activate Blueberry

	// Case: Timestamp is after next validator start time
	// Add a pending validator
	pendingValidatorStartTime := defaultGenesisTime.Add(1 * time.Second)
	pendingValidatorEndTime := pendingValidatorStartTime.Add(defaultMinStakingDuration)
	factory := crypto.FactorySECP256K1R{}
	nodeIDKey, _ := factory.NewPrivateKey()
	rewardAddress := nodeIDKey.PublicKey().Address()
	nodeID := ids.NodeID(rewardAddress)

	_, err := addPendingValidator(
		env,
		pendingValidatorStartTime,
		pendingValidatorEndTime,
		nodeID,
		rewardAddress,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0]},
	)
	assert.NoError(err)

	// add Staker0 (with the right end time) to state
	// so to allow proposalBlk issuance
	staker0StartTime := defaultGenesisTime
	staker0EndTime := pendingValidatorStartTime
	addStaker0, err := env.txBuilder.NewAddValidatorTx(
		10,
		uint64(staker0StartTime.Unix()),
		uint64(staker0EndTime.Unix()),
		ids.GenerateTestNodeID(),
		ids.GenerateTestShortID(),
		reward.PercentDenominator,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	// store Staker0 to state
	staker := state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
	env.state.PutCurrentValidator(staker)
	env.state.AddTx(addStaker0, status.Committed)
	assert.NoError(env.state.Commit())

	// create rewardTx for staker0
	s0RewardTx := &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addStaker0.ID(),
		},
	}
	assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

	// build proposal block moving ahead chain time
	preferredID := env.state.GetLastAccepted()
	parentBlk, _, err := env.state.GetStatelessBlock(preferredID)
	assert.NoError(err)
	statelessProposalBlock, err := stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(pendingValidatorStartTime.Unix()),
		parentBlk.ID(),
		parentBlk.Height()+1,
		s0RewardTx,
	)
	assert.NoError(err)
	propBlk := env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(propBlk.Verify())

	options, err := propBlk.(snowman.OracleBlock).Options()
	assert.NoError(err)
	commitBlk := options[0]
	assert.NoError(commitBlk.Verify())

	assert.NoError(propBlk.Accept())
	assert.NoError(commitBlk.Accept())

	// Test validator weight before delegation
	primarySet, ok := env.config.Validators.GetValidators(constants.PrimaryNetworkID)
	assert.True(ok)
	vdrWeight, _ := primarySet.GetWeight(nodeID)
	assert.Equal(env.config.MinValidatorStake, vdrWeight)

	// Add delegator
	pendingDelegatorStartTime := pendingValidatorStartTime.Add(1 * time.Second)
	pendingDelegatorEndTime := pendingDelegatorStartTime.Add(defaultMinStakingDuration)
	addDelegatorTx, err := env.txBuilder.NewAddDelegatorTx(
		env.config.MinDelegatorStake,
		uint64(pendingDelegatorStartTime.Unix()),
		uint64(pendingDelegatorEndTime.Unix()),
		nodeID,
		preFundedKeys[0].PublicKey().Address(),
		[]*crypto.PrivateKeySECP256K1R{
			preFundedKeys[0],
			preFundedKeys[1],
			preFundedKeys[4],
		},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	staker = state.NewPrimaryNetworkStaker(addDelegatorTx.ID(), &addDelegatorTx.Unsigned.(*txs.AddDelegatorTx).Validator)
	staker.NextTime = staker.StartTime
	staker.Priority = state.PrimaryNetworkDelegatorPendingPriority

	env.state.PutPendingDelegator(staker)
	env.state.AddTx(addDelegatorTx, status.Committed)
	env.state.SetHeight( /*dummyHeight*/ uint64(1))
	assert.NoError(env.state.Commit())

	// add Staker0 (with the right end time) to state
	// so to allow proposalBlk issuance
	staker0EndTime = pendingDelegatorStartTime
	addStaker0, err = env.txBuilder.NewAddValidatorTx(
		10,
		uint64(staker0StartTime.Unix()),
		uint64(staker0EndTime.Unix()),
		ids.GenerateTestNodeID(),
		ids.GenerateTestShortID(),
		reward.PercentDenominator,
		[]*crypto.PrivateKeySECP256K1R{preFundedKeys[0], preFundedKeys[1]},
		ids.ShortEmpty,
	)
	assert.NoError(err)

	// store Staker0 to state
	staker = state.NewPrimaryNetworkStaker(addStaker0.ID(), &addStaker0.Unsigned.(*txs.AddValidatorTx).Validator)
	staker.NextTime = staker.EndTime
	staker.Priority = state.PrimaryNetworkValidatorCurrentPriority
	env.state.PutCurrentValidator(staker)
	env.state.AddTx(addStaker0, status.Committed)
	assert.NoError(env.state.Commit())

	// create rewardTx for staker0
	s0RewardTx = &txs.Tx{
		Unsigned: &txs.RewardValidatorTx{
			TxID: addStaker0.ID(),
		},
	}
	assert.NoError(s0RewardTx.Sign(txs.Codec, nil))

	// Advance Time
	preferredID = env.state.GetLastAccepted()
	parentBlk, _, err = env.state.GetStatelessBlock(preferredID)
	assert.NoError(err)
	statelessProposalBlock, err = stateless.NewProposalBlock(
		version.BlueberryBlockVersion,
		uint64(pendingDelegatorStartTime.Unix()),
		parentBlk.ID(),
		parentBlk.Height()+1,
		s0RewardTx,
	)
	assert.NoError(err)
	propBlk = env.blkManager.NewBlock(statelessProposalBlock)
	assert.NoError(propBlk.Verify())

	options, err = propBlk.(snowman.OracleBlock).Options()
	assert.NoError(err)
	commitBlk = options[0]
	assert.NoError(commitBlk.Verify())

	assert.NoError(propBlk.Accept())
	assert.NoError(commitBlk.Accept())

	// Test validator weight after delegation
	vdrWeight, _ = primarySet.GetWeight(nodeID)
	assert.Equal(env.config.MinDelegatorStake+env.config.MinValidatorStake, vdrWeight)
}
