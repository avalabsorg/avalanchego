// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package mempool

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	transactions "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var _ BlockTimer = &noopBlkTimer{}

type noopBlkTimer struct{}

func (bt *noopBlkTimer) ResetBlockTimer() {}

var preFundedKeys = crypto.BuildTestKeys()

// shows that valid tx is not added to mempool if this would exceed its maximum
// size
func TestBlockBuilderMaxMempoolSizeHandling(t *testing.T) {
	require := require.New(t)

	registerer := prometheus.NewRegistry()
	mpool, err := NewMempool("mempool", registerer, &noopBlkTimer{})
	require.NoError(err)

	txs, err := createTestDecisiontxs(1)
	require.NoError(err)
	tx := txs[0]

	// shortcut to simulated almost filled mempool
	mpool.(*mempool).bytesAvailable = len(tx.Bytes()) - 1

	err = mpool.Add(tx)
	require.True(errors.Is(err, errMempoolFull), err, "max mempool size breached")

	// shortcut to simulated almost filled mempool
	mpool.(*mempool).bytesAvailable = len(tx.Bytes())

	err = mpool.Add(tx)
	require.NoError(err, "should have added tx to mempool")
}

func TestDecisionTxsInMempool(t *testing.T) {
	require := require.New(t)

	registerer := prometheus.NewRegistry()
	mpool, err := NewMempool("mempool", registerer, &noopBlkTimer{})
	require.NoError(err)

	txs, err := createTestDecisiontxs(2)
	require.NoError(err)

	// txs must not already there before we start
	require.False(mpool.HasDecisionTxs())

	for _, tx := range txs {
		// tx not already there
		require.False(mpool.Has(tx.ID()))

		// we can insert
		require.NoError(mpool.Add(tx))

		// we can get it
		require.True(mpool.Has(tx.ID()))

		retrieved := mpool.Get(tx.ID())
		require.True(retrieved != nil)
		require.Equal(tx, retrieved)

		// we can peek it
		peeked := mpool.PeekDecisionTxs(math.MaxInt)

		// tx will be among those peeked,
		// in NO PARTICULAR ORDER
		found := false
		for _, pk := range peeked {
			if pk.ID() == tx.ID() {
				found = true
				break
			}
		}
		require.True(found)

		// once removed it cannot be there
		mpool.Remove([]*transactions.Tx{tx})

		require.False(mpool.Has(tx.ID()))
		require.Equal((*transactions.Tx)(nil), mpool.Get(tx.ID()))

		// we can reinsert it
		require.NoError(mpool.Add(tx))

		// we can mark it as dropped, but it'll still be in mempool
		mpool.MarkDropped(tx.ID(), "dropped for test")
		require.True(mpool.Has(tx.ID()))
		require.Equal(tx, mpool.Get(tx.ID()))
		_, dropped := mpool.GetDropReason(tx.ID())
		require.True(dropped)

		// we can pop it
		txSize := len(tx.Bytes())
		popped := mpool.PopDecisionTxs(math.MaxInt)
		found = false
		for _, pk := range popped {
			if pk.ID() == tx.ID() {
				found = true
				break
			}
		}
		require.True(found)

		// once popped it cannot be there
		require.False(mpool.Has(tx.ID()))
		require.Equal((*transactions.Tx)(nil), mpool.Get(tx.ID()))
		require.Equal([]*transactions.Tx{}, mpool.PeekDecisionTxs(txSize))

		// we can reinsert it again to grow the mempool
		require.NoError(mpool.Add(tx))
	}
}

func TestProposalTxsInMempool(t *testing.T) {
	require := require.New(t)

	registerer := prometheus.NewRegistry()
	mpool, err := NewMempool("mempool", registerer, &noopBlkTimer{})
	require.NoError(err)

	// it's key to this test that proposal txs
	// are ordered by decreasing start time
	txs, err := createTestProposaltxs(2)
	require.NoError(err)

	// txs should not be already there
	require.False(mpool.HasProposalTx())

	for _, tx := range txs {
		require.False(mpool.Has(tx.ID()))

		// we can insert
		require.NoError(mpool.Add(tx))

		// we can get it
		require.True(mpool.HasProposalTx())
		require.True(mpool.Has(tx.ID()))

		retrieved := mpool.Get(tx.ID())
		require.True(retrieved != nil)
		require.Equal(tx, retrieved)

		// we can peek it
		peeked := mpool.PeekProposalTx()
		require.True(peeked != nil)
		require.Equal(tx, peeked)

		// once removed it cannot be there
		mpool.Remove([]*transactions.Tx{tx})

		require.False(mpool.Has(tx.ID()))
		require.Equal((*transactions.Tx)(nil), mpool.Get(tx.ID()))

		// we can reinsert it
		require.NoError(mpool.Add(tx))

		// we can pop it
		popped := mpool.PopProposalTx()
		require.Equal(tx, popped)

		// once popped it cannot be there
		require.False(mpool.Has(tx.ID()))
		require.Equal((*transactions.Tx)(nil), mpool.Get(tx.ID()))

		// we can reinsert it again to grow the mempool
		require.NoError(mpool.Add(tx))
	}
}

func createTestDecisiontxs(count int) ([]*transactions.Tx, error) {
	res := make([]*transactions.Tx, 0, count)
	for i := uint32(0); i < uint32(count); i++ {
		utx := &transactions.CreateChainTx{
			BaseTx: transactions.BaseTx{BaseTx: avax.BaseTx{
				NetworkID:    10,
				BlockchainID: ids.Empty.Prefix(uint64(i)),
				Ins: []*avax.TransferableInput{{
					UTXOID: avax.UTXOID{
						TxID:        ids.ID{'t', 'x', 'I', 'D'},
						OutputIndex: i,
					},
					Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
					In: &secp256k1fx.TransferInput{
						Amt:   uint64(5678),
						Input: secp256k1fx.Input{SigIndices: []uint32{i}},
					},
				}},
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
			}},
			SubnetID:    ids.GenerateTestID(),
			ChainName:   "chainName",
			VMID:        ids.GenerateTestID(),
			FxIDs:       []ids.ID{ids.GenerateTestID()},
			GenesisData: []byte{'g', 'e', 'n', 'D', 'a', 't', 'a'},
			SubnetAuth:  &secp256k1fx.Input{SigIndices: []uint32{1}},
		}

		tx, err := transactions.NewSigned(utx, transactions.Codec, nil)
		if err != nil {
			return nil, err
		}
		res = append(res, tx)
	}
	return res, nil
}

// Proposal txs are sorted by decreasing start time
func createTestProposaltxs(count int) ([]*transactions.Tx, error) {
	var clk mockable.Clock
	res := make([]*transactions.Tx, 0, count)
	for i := 0; i < count; i++ {
		utx := &transactions.AddValidatorTx{
			BaseTx: transactions.BaseTx{},
			Validator: validator.Validator{
				Start: uint64(clk.Time().Add(time.Duration(count-i) * time.Second).Unix()),
			},
			Stake:        nil,
			RewardsOwner: &secp256k1fx.OutputOwners{},
			Shares:       100,
		}

		tx, err := transactions.NewSigned(utx, transactions.Codec, nil)
		if err != nil {
			return nil, err
		}
		res = append(res, tx)
	}
	return res, nil
}
