// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package uptime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMeter(t *testing.T) {
	m := NewMeter(time.Second)
	assert.NotNil(t, m, "should have returned a valid interface")
}

func TestMeter(t *testing.T) {
	halflife := time.Second
	m := &meter{halflife: halflife}

	currentTime := time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC)
	m.clock.Set(currentTime)

	m.Start()

	currentTime = currentTime.Add(halflife)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	m.Start()

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	m.Stop()

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	m.Stop()

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	currentTime = currentTime.Add(halflife)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .25 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .25, uptime)
	}

	m.Start()

	currentTime = currentTime.Add(halflife)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .625 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .625, uptime)
	}
}

func TestMeterTimeTravel(t *testing.T) {
	halflife := time.Second
	m := &meter{
		running: false,
		started: time.Time{},

		halflife: halflife,
		value:    0,
	}

	currentTime := time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC)
	m.clock.Set(currentTime)

	m.lastUpdated = m.clock.Time()

	m.Start()

	currentTime = currentTime.Add(halflife)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	m.Stop()

	currentTime = currentTime.Add(-halflife)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}

	m.Start()

	currentTime = currentTime.Add(halflife / 2)
	m.clock.Set(currentTime)

	if uptime := m.Read(); uptime != .5 {
		t.Fatalf("Wrong uptime value. Expected %f got %f", .5, uptime)
	}
}
