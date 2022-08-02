// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateless

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	_ Block = &ApricotStandardBlock{}
	_ Block = &BlueberryStandardBlock{}
)

// NewStandardBlock assumes [txes] are initialized
func NewStandardBlock(
	version uint16,
	timestamp uint64,
	parentID ids.ID,
	height uint64,
	txes []*txs.Tx,
) (Block, error) {
	switch version {
	case ApricotVersion:
		res := &ApricotStandardBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			Txs: txes,
		}
		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := Block(res)
		bytes, err := Codec.Marshal(ApricotVersion, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}

		return res, res.initialize(ApricotVersion, bytes)

	case BlueberryVersion:
		txsBytes := make([][]byte, len(txes))
		for i, tx := range txes {
			txBytes := tx.Bytes()
			txsBytes[i] = txBytes
		}
		res := &BlueberryStandardBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			TxsBytes: txsBytes,
			Txs:      txes,
		}
		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := Block(res)
		bytes, err := Codec.Marshal(BlueberryVersion, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}
		return res, res.initialize(BlueberryVersion, bytes)

	default:
		return nil, fmt.Errorf("unsupported block version %d", version)
	}
}

type ApricotStandardBlock struct {
	CommonBlock `serialize:"true"`

	Txs []*txs.Tx `serialize:"true" json:"txs"`
}

func (b *ApricotStandardBlock) initialize(version uint16, bytes []byte) error {
	if err := b.CommonBlock.initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	for _, tx := range b.Txs {
		if err := tx.Sign(txs.Codec, nil); err != nil {
			return fmt.Errorf("failed to sign block: %w", err)
		}
	}
	return nil
}

func (b *ApricotStandardBlock) BlockTxs() []*txs.Tx { return b.Txs }

func (b *ApricotStandardBlock) Visit(v Visitor) error {
	return v.ApricotStandardBlock(b)
}

type BlueberryStandardBlock struct {
	CommonBlock `serialize:"true"`

	TxsBytes [][]byte `serialize:"false" blueberry:"true" json:"txs"`

	Txs []*txs.Tx
}

func (b *BlueberryStandardBlock) initialize(version uint16, bytes []byte) error {
	if err := b.CommonBlock.initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if b.Txs == nil {
		b.Txs = make([]*txs.Tx, len(b.TxsBytes))
		for i, txBytes := range b.TxsBytes {
			var tx txs.Tx
			if _, err := txs.Codec.Unmarshal(txBytes, &tx); err != nil {
				return fmt.Errorf("failed unmarshalling tx in blueberry block: %w", err)
			}
			if err := tx.Sign(txs.Codec, nil); err != nil {
				return fmt.Errorf("failed to sign block: %w", err)
			}
			b.Txs[i] = &tx
		}
	}
	return nil
}

func (b *BlueberryStandardBlock) BlockTxs() []*txs.Tx { return b.Txs }

func (b *BlueberryStandardBlock) Visit(v Visitor) error {
	return v.BlueberryStandardBlock(b)
}
