// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package builder

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateful"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs/executor"
)

var _ buildingStrategy = &blueberryStrategy{}

type blueberryStrategy struct {
	*blockBuilder

	// inputs
	parentBlkID ids.ID
	parentState state.Chain
	height      uint64

	// outputs to build blocks
	txes    []*txs.Tx
	blkTime time.Time
}

func (b *blueberryStrategy) hasContent() (bool, error) {
	if err := b.selectBlockContent(); err != nil {
		return false, err
	}

	if len(b.txes) == 0 {
		// blueberry allows empty blocks with non-zero timestamp
		// to move ahead chain time
		return !b.blkTime.IsZero(), nil
	}

	// reinsert txes in mempool before returning
	for _, tx := range b.txes {
		switch tx.Unsigned.(type) {
		case *txs.RewardValidatorTx:
			// nothing to do, these tx is generated
			// just in time, not picked from mempool
		case *txs.AddValidatorTx, *txs.AddDelegatorTx, *txs.AddSubnetValidatorTx:
			b.Mempool.AddProposalTx(tx)
		case *txs.CreateChainTx, *txs.CreateSubnetTx, *txs.ImportTx, *txs.ExportTx:
			b.Mempool.AddDecisionTx(tx)
		default:
			return false, fmt.Errorf("unhandled tx type %T, could not reinsert in mempool", b.txes[0].Unsigned)
		}
	}

	return true, nil
}

func (b *blueberryStrategy) selectBlockContent() error {
	// blkTimestamp is zeroed for Apricot blocks
	// it is tentatively set to chainTime for Blueberry ones
	blkTime := b.parentState.GetTimestamp()

	// try including as many standard txs as possible. No need to advance chain time
	if b.HasDecisionTxs() {
		txs := b.PopDecisionTxs(TargetBlockSize)
		b.txes = txs
		b.blkTime = blkTime
		return nil
	}

	// try rewarding stakers whose staking period ends at current chain time.
	stakerTx, shouldReward, err := b.getStakerToReward(b.parentState)
	if err != nil {
		return fmt.Errorf("could not find next staker to reward %s", err)
	}
	if shouldReward {
		rewardValidatorTx, err := b.txBuilder.NewRewardValidatorTx(stakerTx.ID())
		if err != nil {
			return fmt.Errorf("could not build tx to reward staker %s", err)
		}

		b.txes = []*txs.Tx{rewardValidatorTx}
		b.blkTime = blkTime
		return nil
	}

	// try advancing chain time. It may result in empty blocks
	nextChainTime, shouldAdvanceTime, err := b.getNextChainTime(b.parentState)
	if err != nil {
		return fmt.Errorf("could not retrieve next chain time %s", err)
	}
	if shouldAdvanceTime {
		b.txes = []*txs.Tx{}
		b.blkTime = nextChainTime
		return nil
	}

	// clean out the mempool's transactions with invalid timestamps.
	b.dropTooEarlyMempoolProposalTxs()

	// try including a mempool proposal tx is available.
	if !b.HasProposalTx() {
		b.txExecutorBackend.Ctx.Log.Debug("no pending txs to issue into a block")
		return errNoPendingBlocks
	}

	tx := b.PopProposalTx()

	// if the chain timestamp is too far in the past to issue this transaction
	// but according to local time, it's ready to be issued, then attempt to
	// advance the timestamp, so it can be issued.
	startTime := tx.Unsigned.(txs.StakerTx).StartTime()
	maxChainStartTime := b.parentState.GetTimestamp().Add(executor.MaxFutureStartTime)

	b.txes = []*txs.Tx{tx}
	if startTime.After(maxChainStartTime) {
		now := b.txExecutorBackend.Clk.Time()
		b.blkTime = now
	} else {
		b.blkTime = blkTime // do not change chain time
	}
	return err
}

func (b *blueberryStrategy) build() (snowman.Block, error) {
	blkVersion := uint16(stateless.BlueberryVersion)
	if err := b.selectBlockContent(); err != nil {
		return nil, err
	}

	ctx := b.blockBuilder.txExecutorBackend.Ctx
	if len(b.txes) == 0 {
		// empty standard block are allowed to move chain time head
		return stateful.NewStandardBlock(
			blkVersion,
			uint64(b.blkTime.Unix()),
			b.blkManager,
			ctx,
			b.parentBlkID,
			b.height,
			nil,
		)
	}

	switch b.txes[0].Unsigned.(type) {
	case txs.StakerTx,
		*txs.RewardValidatorTx,
		*txs.AdvanceTimeTx:
		return stateful.NewProposalBlock(
			blkVersion,
			uint64(b.blkTime.Unix()),
			b.blkManager,
			ctx,
			b.parentBlkID,
			b.height,
			b.txes[0],
		)

	case *txs.CreateChainTx,
		*txs.CreateSubnetTx,
		*txs.ImportTx,
		*txs.ExportTx:
		return stateful.NewStandardBlock(
			blkVersion,
			uint64(b.blkTime.Unix()),
			b.blkManager,
			ctx,
			b.parentBlkID,
			b.height,
			b.txes,
		)

	default:
		return nil, fmt.Errorf("unhandled tx type, could not include into a block")
	}
}
