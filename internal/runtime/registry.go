package runtime

import (
	"fmt"
	"sync"
)

// RuntimeName identifies a runtime implementation.
type RuntimeName string

const (
	RuntimeTmux RuntimeName = "tmux"
	RuntimeSDK  RuntimeName = "sdk"
)

// Registry manages available runtimes and provides the active runtime.
type Registry struct {
	mu       sync.RWMutex
	runtimes map[RuntimeName]AgentRuntime
	active   RuntimeName
}

// Global registry instance
var globalRegistry = &Registry{
	runtimes: make(map[RuntimeName]AgentRuntime),
	active:   RuntimeTmux,
}

// Register adds a runtime to the global registry.
func Register(name RuntimeName, rt AgentRuntime) {
	globalRegistry.Register(name, rt)
}

// Get returns a runtime by name from the global registry.
func Get(name RuntimeName) (AgentRuntime, error) {
	return globalRegistry.Get(name)
}

// Active returns the currently active runtime from the global registry.
func Active() AgentRuntime {
	return globalRegistry.Active()
}

// SetActive sets the active runtime by name in the global registry.
func SetActive(name RuntimeName) error {
	return globalRegistry.SetActive(name)
}

// Initialize sets up the default runtimes in the global registry.
func Initialize() {
	globalRegistry.Initialize()
}

// Close closes all registered runtimes in the global registry.
func CloseAll() error {
	return globalRegistry.CloseAll()
}

// Register adds a runtime to the registry.
func (r *Registry) Register(name RuntimeName, rt AgentRuntime) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runtimes[name] = rt
}

// Get returns a runtime by name.
func (r *Registry) Get(name RuntimeName) (AgentRuntime, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rt, ok := r.runtimes[name]
	if !ok {
		return nil, fmt.Errorf("runtime not found: %s", name)
	}
	return rt, nil
}

// Active returns the currently active runtime.
func (r *Registry) Active() AgentRuntime {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runtimes[r.active]
}

// ActiveName returns the name of the currently active runtime.
func (r *Registry) ActiveName() RuntimeName {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

// SetActive sets the active runtime by name.
func (r *Registry) SetActive(name RuntimeName) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.runtimes[name]; !ok {
		return fmt.Errorf("runtime not found: %s", name)
	}
	r.active = name
	return nil
}

// List returns all registered runtime names.
func (r *Registry) List() []RuntimeName {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]RuntimeName, 0, len(r.runtimes))
	for name := range r.runtimes {
		names = append(names, name)
	}
	return names
}

// Initialize sets up the default runtimes.
func (r *Registry) Initialize() {
	r.Register(RuntimeTmux, NewTmuxRuntime())
	// SDK runtime will be registered in Phase 3
}

// CloseAll closes all registered runtimes.
func (r *Registry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for _, rt := range r.runtimes {
		if err := rt.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// NewRegistry creates a new registry instance.
// Use this for testing or when you need multiple registries.
func NewRegistry() *Registry {
	return &Registry{
		runtimes: make(map[RuntimeName]AgentRuntime),
		active:   RuntimeTmux,
	}
}
