package router

import "time"

// HealthConfig describes parameters for router health checks.
type HealthConfig struct {
	// Reports unhealthy if we drop more than [MaxDropRate] of messages
	MaxDropRate float64

	// Halflife of averager used to calculate the message drop rate
	// Must be > 0.
	// Larger value --> Drop rate affected less by recent messages
	MaxDropRateHalflife time.Duration

	// Reports unhealthy if more than this number of requests are outstanding.
	// Must be > 0
	MaxOutstandingRequests int

	// Reports unhealthy if there is at least 1 outstanding request continuously
	// for longer than this
	MaxTimeSinceNoOutstandingRequests time.Duration
}
