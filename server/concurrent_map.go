package server

import (
	"maps"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		mu: sync.RWMutex{},
		m:  make(map[K]V),
	}
}

func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.m[key]
	return value, ok
}

func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.m[key] = value
}

func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.m, key)
}

func (m *ConcurrentMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.m)
}

func (m *ConcurrentMap[K, V]) Range(fn func(key K, value V) bool) {
	m.mu.RLock()
	items := make(map[K]V, len(m.m))
	maps.Copy(items, m.m)
	m.mu.RUnlock()

	for key, value := range items {
		if !fn(key, value) {
			return
		}
	}
}
