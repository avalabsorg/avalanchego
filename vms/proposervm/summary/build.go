// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package summary

import (
	"fmt"

	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
)

func BuildProposerSummary(
	proBlkBytes []byte,
	coreSummary block.StateSummary,
) (ProposerSummary, error) {
	statelessSummary := StatelessSummary{
		ProBlkBytes:  proBlkBytes,
		InnerSummary: coreSummary.Bytes(),
	}

	proSummaryBytes, err := Codec.Marshal(codecVersion, &statelessSummary)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal proposer summary due to: %w", err)
	}
	if err := statelessSummary.initialize(proSummaryBytes); err != nil {
		return nil, err
	}

	return NewProposerSummary(&statelessSummary, coreSummary.Height()), nil
}
