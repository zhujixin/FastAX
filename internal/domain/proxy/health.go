package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// HealthChecker performs periodic health checks on suppliers.
// Results feed into Router.SelectChannel to filter unhealthy channels.
type HealthChecker struct {
	mu       sync.RWMutex
	db       *gorm.DB
	client   *http.Client
	statuses map[uint]string // channelID → "healthy" | "unhealthy" | "unknown"
	interval time.Duration
	stopCh   chan struct{}
}

func NewHealthChecker(db *gorm.DB, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		statuses: make(map[uint]string),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins periodic health checks
func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// checkAll checks all active suppliers
func (hc *HealthChecker) checkAll() {
	var suppliers []model.Supplier
	if err := hc.db.Where("status = ?", 1).Find(&suppliers).Error; err != nil {
		return
	}

	for _, s := range suppliers {
		status := hc.ping(s.APIBaseURL)
		hc.mu.Lock()
		hc.statuses[s.ID] = status
		hc.mu.Unlock()

		// Update provider health record
		hc.recordHealth(s.ID, status)
	}
}

// ping checks if a supplier API is reachable.
// Tries GET to the base URL; considers 2xx/3xx/4xx as healthy (server is up),
// only 5xx or connection errors as unhealthy.
func (hc *HealthChecker) ping(baseURL string) string {
	resp, err := hc.client.Get(baseURL)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	// 5xx = server error → unhealthy
	// 4xx = client error (auth, etc.) → still healthy (server is up)
	// 2xx/3xx = healthy
	if resp.StatusCode >= 500 {
		return "unhealthy"
	}
	return "healthy"
}

func (hc *HealthChecker) recordHealth(channelID uint, status string) {
	now := time.Now()
	health := model.ProviderHealth{
		ProviderID:  channelID,
		Status:      statusToInt(status),
		CheckCount:  1,
		PeriodStart: now.Add(-5 * time.Minute).Unix(),
		PeriodEnd:   now.Unix(),
	}
	hc.db.Create(&health)
}

// GetStatus returns the health status of a channel.
// Returns "healthy", "unhealthy", or "unknown" if never checked.
func (hc *HealthChecker) GetStatus(channelID uint) string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	if status, ok := hc.statuses[channelID]; ok {
		return status
	}
	return "unknown"
}

// SetStatus manually sets the health status of a channel (for testing)
func (hc *HealthChecker) SetStatus(channelID uint, status string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.statuses[channelID] = status
}

// GetAllStatuses returns a copy of all health statuses (for testing/debugging)
func (hc *HealthChecker) GetAllStatuses() map[uint]string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	result := make(map[uint]string, len(hc.statuses))
	for k, v := range hc.statuses {
		result[k] = v
	}
	return result
}

func statusToInt(status string) int {
	switch status {
	case "healthy":
		return 1
	case "unhealthy":
		return 0
	default:
		return -1
	}
}
