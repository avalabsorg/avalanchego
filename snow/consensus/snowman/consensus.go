// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/api/health"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/consensus/snowball"
	"github.com/ava-labs/avalanchego/utils/bag"
)

// Consensus represents a general snowman instance that can be used directly to
// process a series of dependent operations.
type Consensus interface {
	health.Checker

	// Takes in the context, snowball parameters, and the last accepted block.
	Initialize(
		ctx *snow.ConsensusContext,
		params snowball.Parameters,
		lastAcceptedID ids.ID,
		lastAcceptedHeight uint64,
		lastAcceptedTime time.Time,
	) error

	// Returns the number of blocks processing
	NumProcessing() int

	// Adds a new decision. Assumes the dependency has already been added.
	// Returns if a critical error has occurred.
	Add(context.Context, Block) error

	// Decided returns true if the block has been decided.
	Decided(Block) bool

	// Processing returns true if the block ID is currently processing.
	Processing(ids.ID) bool

	// IsPreferred returns true if the block is currently on the preferred
	// chain.
	IsPreferred(Block) bool

	// Returns the ID and height of the last accepted decision.
	LastAccepted() (ids.ID, uint64)

	// Returns the ID of the tail of the strongly preferred sequence of
	// decisions.
	Preference() ids.ID

	// RecordPoll collects the results of a network poll. Assumes all decisions
	// have been previously added. Returns if a critical error has occurred.
	RecordPoll(context.Context, bag.Bag[ids.ID]) error
}
