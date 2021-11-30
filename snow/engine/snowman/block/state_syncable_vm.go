// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"

	"github.com/ava-labs/avalanchego/ids"
)

var ErrStateSyncableVMNotImplemented = errors.New("vm does not implement StateSyncableVM interface")

type Summary struct {
	Key   []byte
	State []byte
}

type StateSyncableVM interface {
	// Enabled indicates whether the state sync is enabled for this VM
	StateSyncEnabled() (bool, error)
	// StateSummary returns latest key and StateSummary with an optional error
	StateSyncGetLastSummary() (Summary, error)
	// IsAccepted returns true if input []bytes represent a valid state summary
	// for fast sync.
	StateSyncIsSummaryAccepted([]byte) (bool, error)

	// SyncState is called with a list of valid state summaries to sync from.
	// These summaries were collected from peers and validated with validators.
	// VM will use information inside the summary to choose one and sync
	// its state to that summary. Normal bootstrapping resumes after this
	// function returns.
	// Will be called with [nil] if no valid state summaries could be found.
	StateSync([]Summary) error

	// StateSyncLastAccepted returns the last accepted block, its height, and
	// an optional error. This is called after the VM notifies the engine
	// of the fast sync operation on the communication channel.
	StateSyncLastAccepted() (ids.ID, uint64, error) // TODO ABENEGIA: remove and handle via notification
}
