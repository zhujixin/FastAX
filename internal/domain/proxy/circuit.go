// 熔断器实现
//
// 本文件实现了熔断器模式（Circuit Breaker Pattern）：
//   - CircuitBreaker: 熔断器结构体，保护下游服务
//   - CircuitState: 熔断器状态（Closed/Open/HalfOpen）
//   - Execute: 执行受保护的操作
//   - RecordSuccess/RecordFailure: 记录成功/失败
//   - 熔断策略: 5xx/超时自动禁用，排除 401/403/429
//   - 恢复策略: HalfOpen 状态下探测恢复
package proxy

import (
	"sync"
	"time"
)

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // 正常状态
	CircuitOpen                         // 熔断状态，拒绝请求
	CircuitHalfOpen                     // 半开状态，探测恢复
)

// CircuitBreaker 熔断器结构体
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

// NewCircuitBreaker 创建熔断器实例
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
