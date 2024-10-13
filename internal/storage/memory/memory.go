package memory

import "sync"

// InMemoryStorage is an in-memory implementation of the Memory interface.
type InMemoryStorage struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// New creates a new instance of InMemoryStorage.
func New() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]interface{}),
	}
}

// Set stores a value associated with a key.
func (s *InMemoryStorage) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value associated with a key.
func (s *InMemoryStorage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Delete removes a value associated with a key.
func (s *InMemoryStorage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return true
	}
	return false
}

// Clear removes all key-value pairs.
func (s *InMemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
}
