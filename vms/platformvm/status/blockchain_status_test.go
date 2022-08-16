// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package status

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockchainStatusJSON(t *testing.T) {
	require := require.New(t)

	statuses := []BlockchainStatus{
		UnknownChain,
		Validating,
		Created,
		Preferred,
		Syncing,
	}
	for _, status := range statuses {
		statusJSON, err := json.Marshal(status)
		require.NoError(err)

		var parsedStatus BlockchainStatus
		err = json.Unmarshal(statusJSON, &parsedStatus)
		require.NoError(err)
		require.Equal(status, parsedStatus)
	}

	{
		status := BlockchainStatus(math.MaxInt32)
		_, err := json.Marshal(status)
		require.Error(err)
	}

	{
		status := Validating
		err := json.Unmarshal([]byte("null"), &status)
		require.NoError(err)
		require.Equal(Validating, status)
	}

	{
		var status BlockchainStatus
		err := json.Unmarshal([]byte(`"not a status"`), &status)
		require.Error(err)
	}
}

func TestBlockchainStatusVerify(t *testing.T) {
	require := require.New(t)

	statuses := []BlockchainStatus{
		UnknownChain,
		Validating,
		Created,
		Preferred,
		Syncing,
	}
	for _, status := range statuses {
		err := status.Verify()
		require.NoError(err, "%s failed verification", status)
	}

	badStatus := BlockchainStatus(math.MaxInt32)
	err := badStatus.Verify()
	require.Error(err, "%s passed verification", badStatus)
}

func TestBlockchainStatusString(t *testing.T) {
	require := require.New(t)

	require.Equal("Unknown", UnknownChain.String())
	require.Equal("Validating", Validating.String())
	require.Equal("Created", Created.String())
	require.Equal("Preferred", Preferred.String())
	require.Equal("Syncing", Syncing.String())
	require.Equal("Dropped", Dropped.String())

	badStatus := BlockchainStatus(math.MaxInt32)
	require.Equal("Invalid blockchain status", badStatus.String())
}
