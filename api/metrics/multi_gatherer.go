// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	dto "github.com/prometheus/client_model/go"
)

// MultiGatherer extends the Gatherer interface by allowing additional gatherers
// to be registered.
type MultiGatherer interface {
	prometheus.Gatherer

	// Register adds the outputs of [gatherer] to the results of future calls to
	// Gather with the provided [name] added to the metrics.
	Register(name string, gatherer prometheus.Gatherer) error
}

// Deprecated: Use NewPrefixGatherer instead.
func NewMultiGatherer() MultiGatherer {
	return NewPrefixGatherer()
}

type multiGatherer struct {
	lock      sync.RWMutex
	names     []string
	gatherers prometheus.Gatherers
}

func (g *multiGatherer) Gather() ([]*dto.MetricFamily, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.gatherers.Gather()
}
