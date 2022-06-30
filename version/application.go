// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package version

import (
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	errDifferentMajor = errors.New("different major version")

	_ fmt.Stringer = &Semantic{}
)

type Application struct {
	Major int `json:"major" yaml:"major"`
	Minor int `json:"minor" yaml:"minor"`
	Patch int `json:"patch" yaml:"patch"`

	str atomic.Value
}

// The only difference here between Application and Semantic is that Application
// prepends "avalanche/" rather than "v".
func (a *Application) String() string {
	strIntf := a.str.Load()
	if strIntf != nil {
		return strIntf.(string)
	}

	str := fmt.Sprintf(
		"avalanche/%d.%d.%d",
		a.Major,
		a.Minor,
		a.Patch,
	)
	a.str.Store(str)
	return str
}

func (a *Application) Compatible(o *Application) error {
	switch {
	case a.Major > o.Major:
		return errDifferentMajor
	default:
		return nil
	}
}

func (a *Application) Before(o *Application) bool {
	return a.Compare(o) < 0
}

// Compare returns a positive number if s > o, 0 if s == o, or a negative number
// if s < o.
func (a *Application) Compare(o *Application) int {
	if a.Major != o.Major {
		return a.Major - o.Major
	}
	if a.Minor != o.Minor {
		return a.Minor - o.Minor
	}
	return a.Patch - o.Patch
}
