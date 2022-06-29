// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm/state"
)

func TestNewExportTx(t *testing.T) {
	h := newTestHelpersCollection()
	h.ctx.Lock.Lock()
	defer func() {
		if err := internalStateShutdown(h); err != nil {
			t.Fatal(err)
		}
	}()

	type test struct {
		description        string
		destinationChainID ids.ID
		sourceKeys         []*crypto.PrivateKeySECP256K1R
		timestamp          time.Time
		shouldErr          bool
		shouldVerify       bool
	}

	sourceKey := preFundedKeys[0]

	tests := []test{
		{
			description:        "P->X export",
			destinationChainID: xChainID,
			sourceKeys:         []*crypto.PrivateKeySECP256K1R{sourceKey},
			timestamp:          defaultValidateStartTime,
			shouldErr:          false,
			shouldVerify:       true,
		},
		{
			description:        "P->C export",
			destinationChainID: cChainID,
			sourceKeys:         []*crypto.PrivateKeySECP256K1R{sourceKey},
			timestamp:          h.cfg.ApricotPhase5Time,
			shouldErr:          false,
			shouldVerify:       true,
		},
	}

	to := ids.GenerateTestShortID()
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert := assert.New(t)
			tx, err := h.txBuilder.NewExportTx(
				defaultBalance-defaultTxFee, // Amount of tokens to export
				tt.destinationChainID,
				to,
				tt.sourceKeys,
				ids.ShortEmpty, // Change address
			)
			if tt.shouldErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			preferredState := h.tState
			fakedState := state.NewDiff(
				preferredState,
				preferredState.CurrentStakers(),
				preferredState.PendingStakers(),
			)
			fakedState.SetTimestamp(tt.timestamp)

			verifier := MempoolTxVerifier{
				Backend:     &h.execBackend,
				ParentState: fakedState,
				Tx:          tx,
			}
			err = tx.Unsigned.Visit(&verifier)
			if tt.shouldVerify {
				assert.NoError(err)
			} else {
				assert.Error(err)
			}
		})
	}
}
