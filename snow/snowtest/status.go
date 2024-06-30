// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowtest

type Status uint32

// [Undecided] means the operation hasn't been decided yet
// [Accepted] means the operation was accepted
// [Rejected] means the operation will never be accepted
const (
	Undecided Status = iota
	Accepted
	Rejected
)
