package idempotency

import (
	"sync"
)

// Store holds successful JSON responses keyed by Idempotency-Key header.
type Store struct {
	mu sync.RWMutex
	m  map[string]entry
}

type entry struct {
	status int
	body   []byte
}

// NewStore creates an empty in-memory store.
func NewStore() *Store {
	return &Store{m: make(map[string]entry)}
}

// Get returns a cached response if present.
func (s *Store) Get(key string) (body []byte, status int, ok bool) {
	if key == "" {
		return nil, 0, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.m[key]
	if !ok {
		return nil, 0, false
	}
	out := append([]byte(nil), e.body...)
	return out, e.status, true
}

// Put stores a successful response for later replays.
func (s *Store) Put(key string, status int, body []byte) {
	if key == "" || len(body) == 0 {
		return
	}
	b := append([]byte(nil), body...)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.m[key]; !exists {
		s.m[key] = entry{status: status, body: b}
	}
}
