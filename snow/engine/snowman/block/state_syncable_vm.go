// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"

	"github.com/ava-labs/avalanchego/ids"
)

var ErrStateSyncableVMNotImplemented = errors.New("vm does not implement StateSyncableVM interface")

// Notes: Key should uniquely identify Summary.
// For Default Network VMs, Key is concatenation of ProposerBlockID + InnerVMBlockID + hash of State
// Key is a string to be easily used as key map
type Summary struct {
	Key   []byte
	State []byte
}

type StateSyncableVM interface {
	// Enabled indicates whether the state sync is enabled for this VM
	StateSyncEnabled() (bool, error)
	// StateSummary returns latest Summary with an optional error
	StateSyncGetLastSummary() (Summary, error)
	// IsAccepted returns true if input []bytes represent a valid state summary
	// for fast sync.
	StateSyncIsSummaryAccepted(key []byte) (bool, error)

	// SyncState is called with a list of valid summaries to sync from.
	// These summaries were collected from peers and validated with validators.
	// VM will use information inside the summary to choose one and sync
	// its state to that summary. Normal bootstrapping resumes after this
	// function returns.
	// Will be called with [nil] if no valid state summaries could be found.
	StateSync([]Summary) error

	// At the end of StateSync process, VM will have rebuilt the state of its blockchain
	// up to a given height. However the block associated with that height may be not known
	// to the VM yet. GetLastSummaryBlockID allows retrival of this block from network
	GetLastSummaryBlockID() (ids.ID, error)
}
