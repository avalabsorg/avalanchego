package builder

import (
	"fmt"

	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateful"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks/stateless"
)

// buildingStrategy defines how to create a versioned block.
// Blocks have different specifications/building instructions as defined by the
// fork that the block exists in.
type buildingStrategy interface {
	// select transactions to be included in block,
	// along with its timestamp.
	selectBlockContent() error

	// builds a versioned snowman.Block
	build() (snowman.Block, error)
}

// Factory method that returns the correct building strategy for the
// current fork.
func (b *blockBuilder) getBuildingStrategy() (buildingStrategy, error) {
	preferred, err := b.Preferred()
	if err != nil {
		return nil, err
	}
	preferredDecision, ok := preferred.(stateful.Decision)
	if !ok {
		// The preferred block should always be a decision block
		return nil, fmt.Errorf("expected Decision block but got %T", preferred)
	}
	preferredState := preferredDecision.OnAccept()

	// select transactions to include and finally build the block
	blkVersion := preferred.ExpectedChildVersion()
	prefBlkID := preferred.ID()
	nextHeight := preferred.Height() + 1

	switch blkVersion {
	case stateless.ApricotVersion:
		return &apricotStrategy{
			blockBuilder: b,
			parentBlkID:  prefBlkID,
			parentState:  preferredState,
			height:       nextHeight,
		}, nil
	case stateless.BlueberryVersion:
		return &blueberryStrategy{
			blockBuilder: b,
			parentBlkID:  prefBlkID,
			parentState:  preferredState,
			height:       nextHeight,
		}, nil
	default:
		return nil, fmt.Errorf("unsupporrted block version %d", blkVersion)
	}
}
