// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/snow/consensus/snowman"
	"github.com/ava-labs/gecko/snow/engine/common/queue"
	"github.com/ava-labs/gecko/utils/logging"
)

type parser struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	vm                      ChainVM
}

func (p *parser) Parse(blkBytes []byte) (queue.Job, error) {
	blk, err := p.vm.ParseBlock(blkBytes)
	if err != nil {
		return nil, err
	}
	return &blockJob{
		log:         p.log,
		numAccepted: p.numAccepted,
		numDropped:  p.numDropped,
		blk:         blk,
	}, nil
}

type blockJob struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	blk                     snowman.Block
}

func (b *blockJob) ID() ids.ID { return b.blk.ID() }
func (b *blockJob) MissingDependencies() ids.Set {
	missing := ids.Set{}
	if parent := b.blk.Parent(); parent.Status() != choices.Accepted {
		missing.Add(parent.ID())
	}
	return missing
}
func (b *blockJob) Execute() error {
	if b.MissingDependencies().Len() != 0 {
		b.numDropped.Inc()
		return errors.New("attempting to accept a block with missing dependencies")
	}
	status := b.blk.Status()
	switch status {
	case choices.Unknown, choices.Rejected:
		b.numDropped.Inc()
		return fmt.Errorf("attempting to execute block with status %s", status)
	case choices.Processing:
		if err := b.blk.Verify(); err != nil {
			b.log.Debug("block %s failed verification during bootstrapping due to %s",
				b.blk.ID(), err)
		}

		b.numAccepted.Inc()
		if err := b.blk.Accept(); err != nil {
			b.log.Debug("block %s failed to accept during bootstrapping due to %s",
				b.blk.ID(), err)
			return fmt.Errorf("failed to accept block in bootstrapping: %w", err)
		}
	}
	return nil
}
func (b *blockJob) Bytes() []byte { return b.blk.Bytes() }
