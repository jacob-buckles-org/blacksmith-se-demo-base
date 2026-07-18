// Package events contains the telemetry event taxonomy. Individual event
// definitions are generated from the taxonomy spec — do not edit
// event_*.go by hand; run `go run ./tools/genevents` from backend/ instead.
package events

import (
	"fmt"
	"sort"
	"sync"
)

// Event is implemented by every generated event type.
type Event interface {
	EventName() string
	Validate() error
	Fields() map[string]any
}

var (
	registryMu sync.RWMutex
	registry   = map[string]func() Event{}
)

func register(name string, factory func() Event) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		panic(fmt.Sprintf("duplicate event registration: %s", name))
	}
	registry[name] = factory
}

// New returns a zero value of the named event type.
func New(name string) (Event, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", name)
	}
	return factory(), nil
}

// Names lists all registered event types, sorted.
func Names() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
