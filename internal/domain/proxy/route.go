package proxy

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// ChannelEntry represents a channel in the routing cache
type ChannelEntry struct {
	ChannelID uint
	Model     string
	Group     string
	Priority  int
	Weight    int
	Status    int // 1=enabled, 3=disabled
}

// Router handles routing decisions with in-memory channel cache
type Router struct {
	mu            sync.RWMutex
	channels      []ChannelEntry
	db            *gorm.DB
	healthChecker *HealthChecker
	stopCh        chan struct{}
}

func NewRouter(db *gorm.DB) *Router {
	return &Router{db: db, stopCh: make(chan struct{})}
}

// SetHealthChecker sets the health checker for route filtering
func (r *Router) SetHealthChecker(hc *HealthChecker) {
	r.healthChecker = hc
}

// StartAutoRefresh starts a goroutine that refreshes the channel cache periodically
func (r *Router) StartAutoRefresh(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.SyncChannelCache()
			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop stops the auto-refresh goroutine
func (r *Router) Stop() {
	close(r.stopCh)
}

// LoadChannels loads all ability entries into memory
func (r *Router) LoadChannels() error {
	var abilities []model.Ability
	if err := r.db.Find(&abilities).Error; err != nil {
		return err
	}

	var suppliers []model.Supplier
	if err := r.db.Find(&suppliers).Error; err != nil {
		return err
	}

	supplierMap := make(map[uint]model.Supplier)
	for _, s := range suppliers {
		supplierMap[s.ID] = s
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.channels = make([]ChannelEntry, 0, len(abilities))
	for _, a := range abilities {
		supplier, ok := supplierMap[a.ChannelID]
		if !ok {
			continue
		}
		if supplier.Status != 1 {
			continue
		}
		r.channels = append(r.channels, ChannelEntry{
			ChannelID: a.ChannelID,
			Model:     a.Model,
			Group:     a.Group,
			Priority:  supplier.Priority,
			Weight:    supplier.Weight,
			Status:    supplier.Status,
		})
	}
	return nil
}

// SelectChannel selects the best channel for a given model and group.
// Strategy: priority groups → within same priority, weighted random.
// Filters out disabled and unhealthy channels.
func (r *Router) SelectChannel(group, model string, disabled map[uint]bool) (*ChannelEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter matching channels
	var candidates []ChannelEntry
	for _, ch := range r.channels {
		if ch.Model != model {
			continue
		}
		if group != "" && ch.Group != group {
			continue
		}
		if ch.Status != 1 {
			continue
		}
		if disabled[ch.ChannelID] {
			continue
		}
		// Health check integration: skip unhealthy channels
		if r.healthChecker != nil && r.healthChecker.GetStatus(ch.ChannelID) == "unhealthy" {
			continue
		}
		candidates = append(candidates, ch)
	}

	if len(candidates) == 0 {
		return nil, errors.New("no available channel for model: " + model)
	}

	// Group by priority
	groups := make(map[int][]ChannelEntry)
	for _, ch := range candidates {
		groups[ch.Priority] = append(groups[ch.Priority], ch)
	}

	// Sort priorities descending
	priorities := make([]int, 0, len(groups))
	for p := range groups {
		priorities = append(priorities, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorities)))

	// Try highest priority group first
	for _, p := range priorities {
		chs := groups[p]
		if len(chs) == 0 {
			continue
		}

		// Weighted random selection within priority group
		totalWeight := 0
		for _, ch := range chs {
			totalWeight += ch.Weight
		}
		if totalWeight == 0 {
			// Equal weight
			idx := rand.Intn(len(chs))
			return &chs[idx], nil
		}

		r := rand.Intn(totalWeight)
		cumulative := 0
		for _, ch := range chs {
			cumulative += ch.Weight
			if r < cumulative {
				return &ch, nil
			}
		}
	}

	return nil, errors.New("no available channel")
}

// SyncChannelCache refreshes the channel cache from DB
func (r *Router) SyncChannelCache() error {
	return r.LoadChannels()
}

// GetChannels returns all channels (for testing)
func (r *Router) GetChannels() []ChannelEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ChannelEntry, len(r.channels))
	copy(result, r.channels)
	return result
}
