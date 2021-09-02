// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
)

func BuildUnsigned(
	parentID ids.ID,
	timestamp time.Time,
	pChainHeight uint64,
	blockBytes []byte,
) (Block, error) {
	block := statelessBlock{
		StatelessBlock: statelessUnsignedBlock{
			ParentID:     parentID,
			Timestamp:    timestamp.Unix(),
			PChainHeight: pChainHeight,
			Certificate:  nil,
			Block:        blockBytes,
		},
		timestamp: timestamp,
	}

	bytes, err := c.Marshal(version, &block)
	if err != nil {
		return nil, err
	}
	block.bytes = bytes

	block.id = hashing.ComputeHash256Array(bytes)
	return &block, nil
}

func Build(
	parentID ids.ID,
	timestamp time.Time,
	pChainHeight uint64,
	cert *x509.Certificate,
	blockBytes []byte,
	key crypto.Signer,
) (Block, error) {
	block := statelessBlock{
		StatelessBlock: statelessUnsignedBlock{
			ParentID:     parentID,
			Timestamp:    timestamp.Unix(),
			PChainHeight: pChainHeight,
			Certificate:  cert.Raw,
			Block:        blockBytes,
		},
		timestamp: timestamp,
		cert:      cert,
		proposer:  hashing.ComputeHash160Array(hashing.ComputeHash256(cert.Raw)),
	}

	unsignedBytes, err := c.Marshal(version, &block.StatelessBlock)
	if err != nil {
		return nil, err
	}

	unsignedHash := hashing.ComputeHash256(unsignedBytes)
	block.Signature, err = key.Sign(rand.Reader, unsignedHash, crypto.SHA256)
	if err != nil {
		return nil, err
	}

	block.bytes, err = c.Marshal(version, &block)
	if err != nil {
		return nil, err
	}

	block.id = hashing.ComputeHash256Array(block.bytes)
	return &block, nil
}
