// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sampler

import (
	"testing"
)

func TestUniformReplacer(t *testing.T) { UniformTest(t, &uniformReplacer{}) }
