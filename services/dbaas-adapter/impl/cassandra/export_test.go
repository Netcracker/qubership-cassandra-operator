package cassandra

import "time"

// SetDropRetryForTest lets the external test package shorten the DROP retry policy.
func SetDropRetryForTest(attempts int, delay time.Duration) {
	dropAttempts = attempts
	dropRetryDelay = delay
}
