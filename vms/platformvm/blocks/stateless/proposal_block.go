// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stateless

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/signed"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions/unsigned"
)

var (
	_ ProposalBlockIntf = &ProposalBlock{}
	_ ProposalBlockIntf = &PostForkProposalBlock{}
)

type ProposalBlockIntf interface {
	CommonBlockIntf

	// ProposalTx returns list of transactions
	// contained in the block
	ProposalTx() *signed.Tx
}

func NewProposalBlock(
	version uint16,
	timestamp uint64,
	parentID ids.ID,
	height uint64,
	tx signed.Tx,
) (ProposalBlockIntf, error) {
	// make sure txs to be included in the block
	// are duly initialized
	if err := tx.Sign(unsigned.Codec, nil); err != nil {
		return nil, fmt.Errorf("failed to sign block: %w", err)
	}

	switch version {
	case PreForkVersion:
		res := &ProposalBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			Tx: tx,
		}

		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := CommonBlockIntf(res)
		bytes, err := Codec.Marshal(version, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}
		return res, res.Initialize(version, bytes)

	case PostForkVersion:
		res := &PostForkProposalBlock{
			CommonBlock: CommonBlock{
				PrntID:       parentID,
				Hght:         height,
				BlkTimestamp: timestamp,
			},
			TxBytes: tx.Bytes(),
			Tx:      tx,
		}

		// We serialize this block as a Block so that it can be deserialized into a
		// Block
		blk := CommonBlockIntf(res)
		bytes, err := Codec.Marshal(version, &blk)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal abort block: %w", err)
		}
		return res, res.Initialize(version, bytes)

	default:
		return nil, fmt.Errorf("unsopported block version %d", version)
	}
}

// As is, this is duplication of atomic block. But let's tolerate some code duplication for now
type ProposalBlock struct {
	CommonBlock `serialize:"true"`

	Tx signed.Tx `serialize:"true" json:"tx"`
}

func (pb *ProposalBlock) Initialize(version uint16, bytes []byte) error {
	if err := pb.CommonBlock.Initialize(version, bytes); err != nil {
		return err
	}

	unsignedBytes, err := unsigned.Codec.Marshal(unsigned.Version, &pb.Tx.Unsigned)
	if err != nil {
		return fmt.Errorf("failed to marshal unsigned tx: %w", err)
	}
	signedBytes, err := unsigned.Codec.Marshal(unsigned.Version, &pb.Tx)
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %w", err)
	}
	pb.Tx.Initialize(unsignedBytes, signedBytes)
	return nil
}

func (pb *ProposalBlock) ProposalTx() *signed.Tx { return &pb.Tx }

type PostForkProposalBlock struct {
	CommonBlock `serialize:"true"`

	TxBytes []byte `serialize:"false" postFork:"true" json:"txs"`

	Tx signed.Tx
}

func (ppb *PostForkProposalBlock) Initialize(version uint16, bytes []byte) error {
	if err := ppb.CommonBlock.Initialize(version, bytes); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	var tx signed.Tx
	_, err := unsigned.Codec.Unmarshal(ppb.TxBytes, &tx)
	if err != nil {
		return fmt.Errorf("failed unmarshalling tx in post fork block: %w", err)
	}
	ppb.Tx = tx
	if err := ppb.Tx.Sign(unsigned.Codec, nil); err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	return nil
}

func (ppb *PostForkProposalBlock) ProposalTx() *signed.Tx { return &ppb.Tx }
