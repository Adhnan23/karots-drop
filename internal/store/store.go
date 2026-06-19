package store

import (
	"errors"
	"sync"
	"time"
)

const defaultTTL = 20 * time.Minute

var ErrStoreFull = errors.New("store is full")

type Item struct {
	Data        []byte
	Filename    string
	ContentType string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Encrypted   bool
}

type Store struct {
	mu       sync.RWMutex
	items    map[string]*Item
	stopCh   chan struct{}
	maxItems int
	ttl      time.Duration
}

func New() *Store {
	return newStore(0, defaultTTL)
}

func NewWithMax(maxItems int) *Store {
	return newStore(maxItems, defaultTTL)
}

func NewWithTTL(ttl time.Duration) *Store {
	return newStore(0, ttl)
}

func NewWithMaxAndTTL(maxItems int, ttl time.Duration) *Store {
	return newStore(maxItems, ttl)
}

func newStore(maxItems int, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	s := &Store{
		items:    make(map[string]*Item),
		stopCh:   make(chan struct{}),
		maxItems: maxItems,
		ttl:      ttl,
	}
	go s.cleanup()
	return s
}

func (s *Store) Stop() {
	close(s.stopCh)
}

func (s *Store) Set(code string, data []byte, filename, contentType string, encrypted bool) error {
	return s.SetWithTTL(code, data, filename, contentType, encrypted, s.ttl)
}

func (s *Store) SetWithTTL(code string, data []byte, filename, contentType string, encrypted bool, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.maxItems > 0 && len(s.items) >= s.maxItems {
		return ErrStoreFull
	}
	now := time.Now()
	s.items[code] = &Item{
		Data:        data,
		Filename:    filename,
		ContentType: contentType,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		Encrypted:   encrypted,
	}
	return nil
}

func (s *Store) Get(code string) (*Item, bool) {
	s.mu.RLock()
	item, ok := s.items[code]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(item.ExpiresAt) {
		s.Delete(code)
		return nil, false
	}
	return item, true
}

func (s *Store) Exists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[code]
	if !ok {
		return false
	}
	return !time.Now().After(item.ExpiresAt)
}

func (s *Store) Delete(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, code)
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *Store) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for code, item := range s.items {
				if now.After(item.ExpiresAt) {
					delete(s.items, code)
				}
			}
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}
