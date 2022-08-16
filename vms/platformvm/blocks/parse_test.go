// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	transactions "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var preFundedKeys = crypto.BuildTestKeys()

func TestStandardBlocks(t *testing.T) {
	// check Apricot standard block can be built and parsed
	require := require.New(t)
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	txs, err := testDecisionTxs()
	require.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		apricotStandardBlk, err := NewApricotStandardBlock(parentID, height, txs)
		require.NoError(err)

		// parse block
		parsedBlueberry, err := Parse(cdc, apricotStandardBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotStandardBlk.ID(), parsedBlueberry.ID())
		require.Equal(apricotStandardBlk.Bytes(), parsedBlueberry.Bytes())
		require.Equal(apricotStandardBlk.Parent(), parsedBlueberry.Parent())
		require.Equal(apricotStandardBlk.Height(), parsedBlueberry.Height())

		// timestamp is not serialized in apricot blocks
		// no matter if block is built with a non-zero timestamp
		require.Equal(time.Unix(0, 0), parsedBlueberry.Timestamp())

		parsedApricot, ok := parsedBlueberry.(*ApricotStandardBlock)
		require.True(ok)
		require.Equal(txs, parsedApricot.Txs())

		// check that blueberry standard block can be built and parsed
		blueberryStandardBlk, err := NewBlueberryStandardBlock(blkTimestamp, parentID, height, txs)
		require.NoError(err)

		// parse block
		parsedBlueberry, err = Parse(cdc, blueberryStandardBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(blueberryStandardBlk.ID(), parsedBlueberry.ID())
		require.Equal(blueberryStandardBlk.Bytes(), parsedBlueberry.Bytes())
		require.Equal(blueberryStandardBlk.Parent(), blueberryStandardBlk.Parent())
		require.Equal(blueberryStandardBlk.Height(), parsedBlueberry.Height())
		require.Equal(blueberryStandardBlk.Timestamp(), parsedBlueberry.Timestamp())
		parsedBlueberryStandardBlk, ok := parsedBlueberry.(*BlueberryStandardBlock)
		require.True(ok)
		require.Equal(txs, parsedBlueberryStandardBlk.Txs())

		// backward compatibility check
		require.Equal(parsedApricot.Txs(), parsedBlueberryStandardBlk.Txs())
	}
}

func TestProposalBlocks(t *testing.T) {
	// check Apricot proposal block can be built and parsed
	require := require.New(t)
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	tx, err := testProposalTx()
	require.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		apricotProposalBlk, err := NewApricotProposalBlock(
			parentID,
			height,
			tx,
		)
		require.NoError(err)

		// parse block
		parsed, err := Parse(cdc, apricotProposalBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotProposalBlk.ID(), parsed.ID())
		require.Equal(apricotProposalBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotProposalBlk.Parent(), parsed.Parent())
		require.Equal(apricotProposalBlk.Height(), parsed.Height())

		// timestamp is not serialized in apricot blocks
		// no matter if block is built with a non-zero timestamp
		require.Equal(time.Unix(0, 0), parsed.Timestamp())

		parsedApricotProposalBlk, ok := parsed.(*ApricotProposalBlock)
		require.True(ok)
		require.Equal([]*transactions.Tx{tx}, parsedApricotProposalBlk.Txs())

		// check that blueberry proposal block can be built and parsed
		blueberryProposalBlk, err := NewBlueberryProposalBlock(
			blkTimestamp,
			parentID,
			height,
			tx,
		)
		require.NoError(err)

		// parse block
		parsed, err = Parse(cdc, blueberryProposalBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(blueberryProposalBlk.ID(), parsed.ID())
		require.Equal(blueberryProposalBlk.Bytes(), parsed.Bytes())
		require.Equal(blueberryProposalBlk.Parent(), blueberryProposalBlk.Parent())
		require.Equal(blueberryProposalBlk.Height(), parsed.Height())
		require.Equal(blueberryProposalBlk.Timestamp(), parsed.Timestamp())
		parsedBlueberryProposalBlk, ok := parsed.(*BlueberryProposalBlock)
		require.True(ok)
		require.Equal([]*transactions.Tx{tx}, parsedBlueberryProposalBlk.Txs())

		// backward compatibility check
		require.Equal(parsedApricotProposalBlk.Txs(), parsedBlueberryProposalBlk.Txs())
	}
}

func TestCommitBlock(t *testing.T) {
	// check Apricot commit block can be built and parsed
	require := require.New(t)
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		apricotCommitBlk, err := NewApricotCommitBlock(parentID, height)
		require.NoError(err)

		// parse block
		parsed, err := Parse(cdc, apricotCommitBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotCommitBlk.ID(), parsed.ID())
		require.Equal(apricotCommitBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotCommitBlk.Parent(), parsed.Parent())
		require.Equal(apricotCommitBlk.Height(), parsed.Height())

		// timestamp is not serialized in apricot blocks
		// no matter if block is built with a non-zero timestamp
		require.Equal(time.Unix(0, 0), parsed.Timestamp())

		// check that blueberry commit block can be built and parsed
		blueberryCommitBlk, err := NewBlueberryCommitBlock(blkTimestamp, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err = Parse(cdc, blueberryCommitBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(blueberryCommitBlk.ID(), parsed.ID())
		require.Equal(blueberryCommitBlk.Bytes(), parsed.Bytes())
		require.Equal(blueberryCommitBlk.Parent(), blueberryCommitBlk.Parent())
		require.Equal(blueberryCommitBlk.Height(), parsed.Height())
		require.Equal(blueberryCommitBlk.Timestamp(), parsed.Timestamp())
	}
}

func TestAbortBlock(t *testing.T) {
	// check Apricot abort block can be built and parsed
	require := require.New(t)
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		apricotAbortBlk, err := NewApricotAbortBlock(parentID, height)
		require.NoError(err)

		// parse block
		parsed, err := Parse(cdc, apricotAbortBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotAbortBlk.ID(), parsed.ID())
		require.Equal(apricotAbortBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotAbortBlk.Parent(), parsed.Parent())
		require.Equal(apricotAbortBlk.Height(), parsed.Height())

		// timestamp is not serialized in apricot blocks
		// no matter if block is built with a non-zero timestamp
		require.Equal(time.Unix(0, 0), parsed.Timestamp())

		// check that blueberry abort block can be built and parsed
		blueberryAbortBlk, err := NewBlueberryAbortBlock(blkTimestamp, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err = Parse(cdc, blueberryAbortBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(blueberryAbortBlk.ID(), parsed.ID())
		require.Equal(blueberryAbortBlk.Bytes(), parsed.Bytes())
		require.Equal(blueberryAbortBlk.Parent(), blueberryAbortBlk.Parent())
		require.Equal(blueberryAbortBlk.Height(), parsed.Height())
		require.Equal(blueberryAbortBlk.Timestamp(), parsed.Timestamp())
	}
}

func TestAtomicBlock(t *testing.T) {
	// check atomic block can be built and parsed
	require := require.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	tx, err := testAtomicTx()
	require.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		atomicBlk, err := NewApricotAtomicBlock(
			parentID,
			height,
			tx,
		)
		require.NoError(err)

		// parse block
		parsed, err := Parse(cdc, atomicBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(atomicBlk.ID(), parsed.ID())
		require.Equal(atomicBlk.Bytes(), parsed.Bytes())
		require.Equal(atomicBlk.Parent(), parsed.Parent())
		require.Equal(atomicBlk.Height(), parsed.Height())
		require.Equal(time.Unix(0, 0), parsed.Timestamp())

		parsedAtomicBlk, ok := parsed.(*ApricotAtomicBlock)
		require.True(ok)
		require.Equal([]*transactions.Tx{tx}, parsedAtomicBlk.Txs())
	}
}

func testAtomicTx() (*transactions.Tx, error) {
	utx := &transactions.ImportTx{
		BaseTx: transactions.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    10,
			BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
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
			Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}},
		SourceChain: ids.ID{'c', 'h', 'a', 'i', 'n'},
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(1),
				OutputIndex: 1,
			},
			Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
			In: &secp256k1fx.TransferInput{
				Amt:   50000,
				Input: secp256k1fx.Input{SigIndices: []uint32{0}},
			},
		}},
	}
	signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
	return transactions.NewSigned(utx, transactions.Codec, signers)
}

func testDecisionTxs() ([]*transactions.Tx, error) {
	countTxs := 2
	txs := make([]*transactions.Tx, 0, countTxs)
	for i := 0; i < countTxs; i++ {
		// Create the tx
		utx := &transactions.CreateChainTx{
			BaseTx: transactions.BaseTx{BaseTx: avax.BaseTx{
				NetworkID:    10,
				BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
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
				Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			}},
			SubnetID:    ids.ID{'s', 'u', 'b', 'n', 'e', 't', 'I', 'D'},
			ChainName:   "a chain",
			VMID:        ids.GenerateTestID(),
			FxIDs:       []ids.ID{ids.GenerateTestID()},
			GenesisData: []byte{'g', 'e', 'n', 'D', 'a', 't', 'a'},
			SubnetAuth:  &secp256k1fx.Input{SigIndices: []uint32{1}},
		}

		signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
		tx, err := transactions.NewSigned(utx, transactions.Codec, signers)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func testProposalTx() (*transactions.Tx, error) {
	utx := &transactions.RewardValidatorTx{
		TxID: ids.ID{'r', 'e', 'w', 'a', 'r', 'd', 'I', 'D'},
	}

	signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
	return transactions.NewSigned(utx, transactions.Codec, signers)
}
