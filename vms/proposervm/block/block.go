// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"crypto/x509"
	"errors"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

var errUnexpectedSignature = errors.New("expected no signature but one was provided")

type Block interface {
	ID() ids.ID
	ParentID() ids.ID
	Block() []byte
	Bytes() []byte

	initialize(bytes []byte) error
}

type SignedBlock interface {
	Block

	PChainHeight() uint64
	Timestamp() time.Time
	Proposer() ids.ShortID

	Verify(chainID ids.ID) error
}

type statelessUnsignedBlock struct {
	ParentID     ids.ID `serialize:"true"`
	Timestamp    int64  `serialize:"true"`
	PChainHeight uint64 `serialize:"true"`
	Certificate  []byte `serialize:"true"`
	Block        []byte `serialize:"true"`
}

type statelessBlock struct {
	StatelessBlock statelessUnsignedBlock `serialize:"true"`
	Signature      []byte                 `serialize:"true"`

	id        ids.ID
	timestamp time.Time
	cert      *x509.Certificate
	proposer  ids.ShortID
	bytes     []byte
}

func (b *statelessBlock) ID() ids.ID       { return b.id }
func (b *statelessBlock) ParentID() ids.ID { return b.StatelessBlock.ParentID }
func (b *statelessBlock) Block() []byte    { return b.StatelessBlock.Block }
func (b *statelessBlock) Bytes() []byte    { return b.bytes }

func (b *statelessBlock) initialize(bytes []byte) error {
	b.bytes = bytes
	unsignedBytes := bytes[:len(bytes)-wrappers.IntLen-len(b.Signature)]
	b.id = hashing.ComputeHash256Array(unsignedBytes)

	b.timestamp = time.Unix(b.StatelessBlock.Timestamp, 0)
	if len(b.StatelessBlock.Certificate) == 0 {
		return nil
	}

	cert, err := x509.ParseCertificate(b.StatelessBlock.Certificate)
	if err != nil {
		return err
	}
	b.cert = cert
	b.proposer = hashing.ComputeHash160Array(hashing.ComputeHash256(cert.Raw))
	return nil
}

func (b *statelessBlock) PChainHeight() uint64  { return b.StatelessBlock.PChainHeight }
func (b *statelessBlock) Timestamp() time.Time  { return b.timestamp }
func (b *statelessBlock) Proposer() ids.ShortID { return b.proposer }

func (b *statelessBlock) Verify(chainID ids.ID) error {
	if b.cert == nil {
		if len(b.Signature) != 0 {
			return errUnexpectedSignature
		}
		return nil
	}

	header, err := BuildHeader(chainID, b.StatelessBlock.ParentID, b.id)
	if err != nil {
		return err
	}

	headerBytes := header.Bytes()
	return b.cert.CheckSignature(b.cert.SignatureAlgorithm, headerBytes, b.Signature)
}
