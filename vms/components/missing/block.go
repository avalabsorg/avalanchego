// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package missing

import (
	"errors"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/snow/consensus/snowman"
)

var (
	errMissingBlock = errors.New("missing block")
)

// Block represents a block that can't be found
type Block struct{ BlkID ids.ID }

// ID ...
func (mb *Block) ID() ids.ID { return mb.BlkID }

// Accept ...
func (*Block) Accept() error { return errMissingBlock }

// Reject ...
func (*Block) Reject() error { return errMissingBlock }

// Status ...
func (*Block) Status() choices.Status { return choices.Unknown }

// Parent ...
func (*Block) Parent() snowman.Block { return nil }

// Verify ...
func (*Block) Verify() error { return errMissingBlock }

// Bytes ...
func (*Block) Bytes() []byte { return nil }
