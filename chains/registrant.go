// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"github.com/ava-labs/avalanchego/snow"
)

// Registrant can register the existence of a chain
type Registrant interface {
	// If RegisterChain grabs [engine]'s lock, it must do so in a goroutine
	RegisterChain(name string, ctx *snow.Context, engine interface{})
}
