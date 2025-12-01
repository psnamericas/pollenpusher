package format

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// registry holds all registered CDR formats
var (
	registry = make(map[string]CDRFormat)
	mu       sync.RWMutex
)

// Register adds a new format to the registry.
// This is typically called from init() functions in format packages.
func Register(format CDRFormat) error {
	mu.Lock()
	defer mu.Unlock()

	name := strings.ToLower(format.Name())
	if _, exists := registry[name]; exists {
		return fmt.Errorf("format %q already registered", name)
	}

	registry[name] = format
	return nil
}

// MustRegister registers a format and panics on error.
// This is useful for init() functions.
func MustRegister(format CDRFormat) {
	if err := Register(format); err != nil {
		panic(err)
	}
}

// Get retrieves a format by name (case-insensitive)
func Get(name string) (CDRFormat, error) {
	mu.RLock()
	defer mu.RUnlock()

	format, exists := registry[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return format, nil
}

// List returns all registered format names in alphabetical order
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered formats
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry)
}

// ForEach calls the provided function for each registered format
func ForEach(fn func(name string, format CDRFormat)) {
	mu.RLock()
	defer mu.RUnlock()

	for name, format := range registry {
		fn(name, format)
	}
}
