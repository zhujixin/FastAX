package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const pluginTimeout = 500 * time.Millisecond

// Plugin defines the interface for platform plugins
type Plugin interface {
	Name() string
	Init(config map[string]string) error
	OnRequest(req *RequestContext) error
	OnResponse(resp *ResponseContext) error
	OnError(errCtx *ErrorContext)
}

type RequestContext struct {
	TraceID  string
	UserID   uint
	Model    string
	Body     []byte
	Metadata map[string]string
}

type ResponseContext struct {
	TraceID    string
	StatusCode int
	Body       []byte
	LatencyMs  int
}

type ErrorContext struct {
	TraceID string
	Error   error
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	configs map[string]map[string]string
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]map[string]string),
	}
}

func (pm *PluginManager) Register(name string, p Plugin, config map[string]string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	if err := p.Init(config); err != nil {
		return fmt.Errorf("init plugin %s: %w", name, err)
	}

	pm.plugins[name] = p
	pm.configs[name] = config
	return nil
}

func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.plugins[name]; !exists {
		return errors.New("plugin not found")
	}

	delete(pm.plugins, name)
	delete(pm.configs, name)
	return nil
}

func (pm *PluginManager) ExecuteRequest(ctx context.Context, req *RequestContext) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, pluginTimeout)
	defer cancel()

	for _, p := range pm.plugins {
		errCh := make(chan error, 1)
		go func(plug Plugin) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("plugin %s panic: %v", plug.Name(), r)
				}
			}()
			errCh <- plug.OnRequest(req)
		}(p)

		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-timeoutCtx.Done():
			return fmt.Errorf("plugin %s timeout", p.Name())
		}
	}
	return nil
}

func (pm *PluginManager) ExecuteResponse(ctx context.Context, resp *ResponseContext) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, pluginTimeout)
	defer cancel()

	for _, p := range pm.plugins {
		errCh := make(chan error, 1)
		go func(plug Plugin) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("plugin %s panic: %v", plug.Name(), r)
				}
			}()
			errCh <- plug.OnResponse(resp)
		}(p)

		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-timeoutCtx.Done():
			return fmt.Errorf("plugin %s timeout", p.Name())
		}
	}
	return nil
}

func (pm *PluginManager) ExecuteError(errCtx *ErrorContext) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.plugins {
		func() {
			defer func() {
				recover() // swallow panic in error handlers
			}()
			p.OnError(errCtx)
		}()
	}
}

func (pm *PluginManager) List() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

func (pm *PluginManager) Get(name string) Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.plugins[name]
}
