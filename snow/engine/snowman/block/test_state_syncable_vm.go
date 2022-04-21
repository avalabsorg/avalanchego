// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"
	"testing"

	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
)

var (
	_ StateSyncableVM = &TestStateSyncableVM{}

	errStateSyncEnabled           = errors.New("unexpectedly called StateSyncEnabled")
	errGetLastStateSummary        = errors.New("unexpectedly called GetLastStateSummary")
	errParseStateSummary          = errors.New("unexpectedly called ParseStateSummary")
	errGetStateSummary            = errors.New("unexpectedly called GetStateSummary")
	errStateSyncGetOngoingSummary = errors.New("unexpectedly called StateSyncGetOngoingSummary")
	errGetStateSyncResult         = errors.New("unexpectedly called GetStateSyncResult")
	errParseStateSyncableBlock    = errors.New("unexpectedly called ParseStateSyncableBlock")
)

type TestStateSyncableVM struct {
	T *testing.T

	CantStateSyncEnabled, CantStateSyncGetOngoingSummary,
	CantGetLastStateSummary, CantParseStateSummary,
	CantGetStateSummary, CantGetStateSyncResult,
	CantParseStateSyncableBlock bool

	StateSyncEnabledF           func() (bool, error)
	GetOngoingSyncStateSummaryF func() (Summary, error)
	GetLastStateSummaryF        func() (Summary, error)
	ParseStateSummaryF          func(summaryBytes []byte) (Summary, error)
	GetStateSummaryF            func(uint64) (Summary, error)
	GetStateSyncResultF         func() error
	ParseStateSyncableBlockF    func(blkBytes []byte) (snowman.StateSyncableBlock, error)
}

func (tss *TestStateSyncableVM) StateSyncEnabled() (bool, error) {
	if tss.StateSyncEnabledF != nil {
		return tss.StateSyncEnabledF()
	}
	if tss.CantStateSyncEnabled && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called StateSyncEnabled")
	}
	return false, errStateSyncEnabled
}

func (tss *TestStateSyncableVM) GetOngoingSyncStateSummary() (Summary, error) {
	if tss.GetOngoingSyncStateSummaryF != nil {
		return tss.GetOngoingSyncStateSummaryF()
	}
	if tss.CantStateSyncGetOngoingSummary && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called StateSyncGetOngoingSummary")
	}
	return nil, errStateSyncGetOngoingSummary
}

func (tss *TestStateSyncableVM) GetLastStateSummary() (Summary, error) {
	if tss.GetLastStateSummaryF != nil {
		return tss.GetLastStateSummaryF()
	}
	if tss.CantGetLastStateSummary && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called GetLastStateSummary")
	}
	return nil, errGetLastStateSummary
}

func (tss *TestStateSyncableVM) ParseStateSummary(summaryBytes []byte) (Summary, error) {
	if tss.ParseStateSummaryF != nil {
		return tss.ParseStateSummaryF(summaryBytes)
	}
	if tss.CantParseStateSummary && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called ParseStateSummary")
	}
	return nil, errParseStateSummary
}

func (tss *TestStateSyncableVM) GetStateSummary(key uint64) (Summary, error) {
	if tss.GetStateSummaryF != nil {
		return tss.GetStateSummaryF(key)
	}
	if tss.CantGetStateSummary && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called GetStateSummary")
	}
	return nil, errGetStateSummary
}

func (tss *TestStateSyncableVM) GetStateSyncResult() error {
	if tss.GetStateSyncResultF != nil {
		return tss.GetStateSyncResultF()
	}
	if tss.CantGetStateSyncResult && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called GetStateSyncResult")
	}
	return errGetStateSyncResult
}

func (tss *TestStateSyncableVM) ParseStateSyncableBlock(blkBytes []byte) (snowman.StateSyncableBlock, error) {
	if tss.ParseStateSyncableBlockF != nil {
		return tss.ParseStateSyncableBlockF(blkBytes)
	}
	if tss.CantParseStateSyncableBlock && tss.T != nil {
		tss.T.Fatalf("Unexpectedly called ParseStateSyncableBlock")
	}
	return nil, errParseStateSyncableBlock
}
