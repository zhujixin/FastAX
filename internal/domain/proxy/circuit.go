package proxy

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Tripped, rejecting requests
	CircuitHalfOpen                     // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	failureCount    int
	successCount    int
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	state           CircuitState
	lastFailure     time.Time
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		state:            CircuitClosed,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.failureCount = 0
		}
	case CircuitClosed:
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.lastFailure = time.Now()
		}
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		cb.lastFailure = time.Now()
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ShouldDisableChannel determines if a channel should be disabled based on error type
// Returns true for server errors (5xx) and timeouts
// Returns false for client errors (401, 403, 429) - these don't indicate channel failure
func ShouldDisableChannel(statusCode int) bool {
	return statusCode >= 500
}
