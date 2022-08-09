// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metrics

import (
	"net/http"
	"time"

	"github.com/gorilla/rpc/v2"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks"
)

var Noop Metrics = noopMetrics{}

type noopMetrics struct{}

func (noopMetrics) MarkOptionVoteWon() {}

func (noopMetrics) MarkOptionVoteLost() {}

func (noopMetrics) MarkAccepted(blocks.Block) error { return nil }

func (noopMetrics) InterceptRequestFunc() func(*rpc.RequestInfo) *http.Request {
	return func(*rpc.RequestInfo) *http.Request { return nil }
}

func (noopMetrics) AfterRequestFunc() func(*rpc.RequestInfo) {
	return func(ri *rpc.RequestInfo) {}
}

func (noopMetrics) IncValidatorSetsCreated() {}

func (noopMetrics) IncValidatorSetsCached() {}

func (noopMetrics) AddValidatorSetsDuration(time.Duration) {}

func (noopMetrics) AddValidatorSetsHeightDiff(float64) {}

func (noopMetrics) SetLocalStake(float64) {}

func (noopMetrics) SetTotalStake(float64) {}

func (noopMetrics) SetSubnetPercentConnected(ids.ID, float64) {}

func (noopMetrics) SetPercentConnected(float64) {}
