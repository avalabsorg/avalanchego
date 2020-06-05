// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/snow/consensus/avalanche"
	"github.com/ava-labs/gecko/snow/engine/common/queue"
	"github.com/ava-labs/gecko/utils/logging"
)

type vtxParser struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	state                   State
}

func (p *vtxParser) Parse(vtxBytes []byte) (queue.Job, error) {
	vtx, err := p.state.ParseVertex(vtxBytes)
	if err != nil {
		return nil, err
	}
	return &vertexJob{
		log:         p.log,
		numAccepted: p.numAccepted,
		numDropped:  p.numDropped,
		vtx:         vtx,
	}, nil
}

type vertexJob struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	vtx                     avalanche.Vertex
}

func (v *vertexJob) ID() ids.ID { return v.vtx.ID() }
func (v *vertexJob) MissingDependencies() ids.Set {
	missing := ids.Set{}
	for _, parent := range v.vtx.Parents() {
		if parent.Status() != choices.Accepted {
			missing.Add(parent.ID())
		}
	}
	return missing
}
func (v *vertexJob) Execute() error {
	if v.MissingDependencies().Len() != 0 {
		v.numDropped.Inc()
		return errors.New("attempting to execute blocked vertex")
	}
	for _, tx := range v.vtx.Txs() {
		if tx.Status() != choices.Accepted {
			v.numDropped.Inc()
			v.log.Warn("attempting to execute vertex with non-accepted transactions")
			return nil
		}
	}
	status := v.vtx.Status()
	switch status {
	case choices.Unknown, choices.Rejected:
		v.numDropped.Inc()
		return fmt.Errorf("attempting to execute vertex with status %s", status)
	case choices.Processing:
		v.numAccepted.Inc()
		if err := v.vtx.Accept(); err != nil {
			return fmt.Errorf("failed to accept vertex in bootstrapping: %w", err)
		}
	}
	return nil
}
func (v *vertexJob) Bytes() []byte { return v.vtx.Bytes() }
