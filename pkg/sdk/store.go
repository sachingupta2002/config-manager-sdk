package sdk

import (
	"sync"
)

// Store is an in-memory cache for config values
type Store struct {
	mu     sync.RWMutex
	values map[string]ConfigValue
}

// NewStore creates a new Store instance
func NewStore() *Store {
	return &Store{
		values: make(map[string]ConfigValue),
	}
}

// Get retrieves a value from the store
func (s *Store) Get(key string) (ConfigValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}

// Set stores a value in the store
func (s *Store) Set(key string, value ConfigValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

// SetAll replaces all values in the store
func (s *Store) SetAll(values map[string]ConfigValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = values
}

// Delete removes a value from the store
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
}

// Clear removes all values from the store
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]ConfigValue)
}

// All returns a copy of all values
func (s *Store) All() map[string]ConfigValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]ConfigValue, len(s.values))
	for k, v := range s.values {
		copy[k] = v
	}
	return copy
}
