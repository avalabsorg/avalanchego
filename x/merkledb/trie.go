// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package merkledb

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
)

var errNoNewRoot = errors.New("there was no updated root in change list")

type ReadOnlyTrie interface {
	// get the value associated with the key
	// database.ErrNotFound if the key is not present
	GetValue(ctx context.Context, key []byte) ([]byte, error)

	// get the values associated with the keys
	// database.ErrNotFound if the key is not present
	GetValues(ctx context.Context, keys [][]byte) ([][]byte, []error)

	// get the value associated with the key in path form
	// database.ErrNotFound if the key is not present
	getValue(key path) ([]byte, error)

	// get the merkle root of the Trie
	GetMerkleRoot(ctx context.Context) (ids.ID, error)

	// get an editable copy of the node with the given key path
	getEditableNode(ctx context.Context, key path) (*node, error)

	// generate a proof of the value associated with a particular key, or a proof of its absence from the trie
	GetProof(ctx context.Context, bytesPath []byte) (*Proof, error)

	// generate a proof of up to maxLength smallest key/values with keys between start and end
	GetRangeProof(ctx context.Context, start, end []byte, maxLength int) (*RangeProof, error)

	// GetKeyValues but doesn't grab any locks.
	getKeyValues(
		start []byte,
		end []byte,
		maxLength int,
		keysToIgnore set.Set[string],
		lock bool,
	) ([]KeyValue, error)
}

type Trie interface {
	ReadOnlyTrie

	// Delete a key from the Trie
	Remove(ctx context.Context, key []byte) error

	// Get a new view on top of this Trie
	NewPreallocatedView(estimatedChanges int) (TrieView, error)

	// Get a new view on top of this Trie
	NewView() (TrieView, error)

	// Insert a key/value pair into the Trie
	Insert(ctx context.Context, key, value []byte) error
}

type TrieView interface {
	Trie

	// Commits changes in the trieToCommit into the current trie
	// then commits the combined changes down the stack until all changes in the stack commit to the database
	// Takes the DB commit lock
	CommitToDB(ctx context.Context) error

	// commits changes in the trieToCommit into the current trie
	// then commits the combined changes down the stack until all changes in the stack commit to the database
	commitToDB(ctx context.Context, trieToCommit *trieView) error
}
