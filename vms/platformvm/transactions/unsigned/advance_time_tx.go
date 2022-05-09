// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package unsigned

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

var _ Tx = &AdvanceTimeTx{}

// AdvanceTimeTx is a transaction to increase the chain's timestamp.
// When the chain's timestamp is updated (a AdvanceTimeTx is accepted and
// followed by a commit block) the staker set is also updated accordingly.
// It must be that:
//   * proposed timestamp > [current chain time]
//   * proposed timestamp <= [time for next staker set change]
type AdvanceTimeTx struct {
	avax.Metadata

	// Unix time this block proposes increasing the timestamp to
	Time uint64 `serialize:"true" json:"time"`
}

func (tx *AdvanceTimeTx) InitCtx(*snow.Context) {}

// Timestamp returns the time this block is proposing the chain should be set to
func (tx *AdvanceTimeTx) Timestamp() time.Time {
	return time.Unix(int64(tx.Time), 0)
}

func (tx *AdvanceTimeTx) InputIDs() ids.Set {
	return nil
}

func (tx *AdvanceTimeTx) SyntacticVerify(*snow.Context) error {
	return nil
}
