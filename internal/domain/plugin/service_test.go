package plugin

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockPlugin struct {
	name        string
	initErr     error
	onReqErr    error
	onRespErr   error
	onReqCalled bool
	onRespCalled bool
	onErrCalled bool
}

func (m *mockPlugin) Name() string { return m.name }
func (m *mockPlugin) Init(config map[string]string) error { return m.initErr }
func (m *mockPlugin) OnRequest(req *RequestContext) error {
	m.onReqCalled = true
	return m.onReqErr
}
func (m *mockPlugin) OnResponse(resp *ResponseContext) error {
	m.onRespCalled = true
	return m.onRespErr
}
func (m *mockPlugin) OnError(errCtx *ErrorContext) {
	m.onErrCalled = true
}

func TestPluginManager_Register(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}

	err := pm.Register("test", p, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	names := pm.List()
	if len(names) != 1 {
		t.Errorf("len = %v, want 1", len(names))
	}
}

func TestPluginManager_Register_Duplicate(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}

	pm.Register("test", p, nil)
	err := pm.Register("test", p, nil)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestPluginManager_Register_InitError(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test", initErr: errors.New("init failed")}

	err := pm.Register("test", p, nil)
	if err == nil {
		t.Fatal("expected error for init failure")
	}
}

func TestPluginManager_Unregister(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}

	pm.Register("test", p, nil)
	err := pm.Unregister("test")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	if len(pm.List()) != 0 {
		t.Errorf("len = %v, want 0", len(pm.List()))
	}
}

func TestPluginManager_ExecuteRequest(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}
	pm.Register("test", p, nil)

	err := pm.ExecuteRequest(context.Background(), &RequestContext{TraceID: "t1"})
	if err != nil {
		t.Fatalf("ExecuteRequest() error = %v", err)
	}
	if !p.onReqCalled {
		t.Error("OnRequest not called")
	}
}

func TestPluginManager_ExecuteRequest_Error(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test", onReqErr: errors.New("blocked")}
	pm.Register("test", p, nil)

	err := pm.ExecuteRequest(context.Background(), &RequestContext{TraceID: "t1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPluginManager_ExecuteResponse(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}
	pm.Register("test", p, nil)

	pm.ExecuteResponse(context.Background(), &ResponseContext{TraceID: "t1", StatusCode: 200})
	if !p.onRespCalled {
		t.Error("OnResponse not called")
	}
}

func TestPluginManager_ExecuteRequest_Timeout(t *testing.T) {
	pm := NewPluginManager()
	p := &slowPlugin{name: "slow", delay: 2 * time.Second}
	pm.Register("slow", p, nil)

	err := pm.ExecuteRequest(context.Background(), &RequestContext{TraceID: "t1"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "plugin slow timeout" {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestPluginManager_ExecuteResponse_Timeout(t *testing.T) {
	pm := NewPluginManager()
	p := &slowPlugin{name: "slow", delay: 2 * time.Second}
	pm.Register("slow", p, nil)

	err := pm.ExecuteResponse(context.Background(), &ResponseContext{TraceID: "t1", StatusCode: 200})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "plugin slow timeout" {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestPluginManager_ExecuteRequest_Panic(t *testing.T) {
	pm := NewPluginManager()
	p := &panicPlugin{name: "panicker"}
	pm.Register("panicker", p, nil)

	err := pm.ExecuteRequest(context.Background(), &RequestContext{TraceID: "t1"})
	if err == nil {
		t.Fatal("expected panic error")
	}
	if err.Error() != "plugin panicker panic: boom" {
		t.Errorf("expected panic error, got: %v", err)
	}
}

func TestPluginManager_ExecuteResponse_Panic(t *testing.T) {
	pm := NewPluginManager()
	p := &panicPlugin{name: "panicker"}
	pm.Register("panicker", p, nil)

	err := pm.ExecuteResponse(context.Background(), &ResponseContext{TraceID: "t1", StatusCode: 200})
	if err == nil {
		t.Fatal("expected panic error")
	}
	if err.Error() != "plugin panicker panic: boom" {
		t.Errorf("expected panic error, got: %v", err)
	}
}

func TestPluginManager_ExecuteError_Panic(t *testing.T) {
	pm := NewPluginManager()
	p := &panicPlugin{name: "panicker"}
	pm.Register("panicker", p, nil)

	// Should not panic even when plugin panics
	pm.ExecuteError(&ErrorContext{TraceID: "t1", Error: errors.New("test")})
}

func TestPluginManager_ExecuteError(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}
	pm.Register("test", p, nil)

	pm.ExecuteError(&ErrorContext{TraceID: "t1", Error: errors.New("test")})
	if !p.onErrCalled {
		t.Error("OnError not called")
	}
}

func TestPluginManager_Get(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "test"}
	pm.Register("test", p, nil)

	got := pm.Get("test")
	if got == nil {
		t.Fatal("expected plugin")
	}
	if got.Name() != "test" {
		t.Errorf("name = %v", got.Name())
	}
}

// --- Test helpers: slowPlugin and panicPlugin ---

type slowPlugin struct {
	name  string
	delay time.Duration
}

func (s *slowPlugin) Name() string                        { return s.name }
func (s *slowPlugin) Init(config map[string]string) error  { return nil }
func (s *slowPlugin) OnRequest(req *RequestContext) error {
	time.Sleep(s.delay)
	return nil
}
func (s *slowPlugin) OnResponse(resp *ResponseContext) error {
	time.Sleep(s.delay)
	return nil
}
func (s *slowPlugin) OnError(errCtx *ErrorContext) {
	time.Sleep(s.delay)
}

type panicPlugin struct {
	name string
}

func (p *panicPlugin) Name() string                        { return p.name }
func (p *panicPlugin) Init(config map[string]string) error  { return nil }
func (p *panicPlugin) OnRequest(req *RequestContext) error  { panic("boom") }
func (p *panicPlugin) OnResponse(resp *ResponseContext) error { panic("boom") }
func (p *panicPlugin) OnError(errCtx *ErrorContext)         { panic("boom") }
