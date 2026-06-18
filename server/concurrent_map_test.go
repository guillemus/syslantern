package server

import (
	"sync"
	"testing"
)

func TestConcurrentMap(t *testing.T) {
	m := NewConcurrentMap[string, int]()

	m.Set("one", 1)
	value, ok := m.Get("one")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != 1 {
		t.Fatalf("expected value 1, got %d", value)
	}

	if m.Len() != 1 {
		t.Fatalf("expected len 1, got %d", m.Len())
	}

	calls := 0
	m.Range(func(key string, value int) bool {
		calls++
		if key != "one" {
			t.Fatalf("expected key one, got %s", key)
		}
		if value != 1 {
			t.Fatalf("expected value 1, got %d", value)
		}
		return true
	})
	if calls != 1 {
		t.Fatalf("expected 1 range call, got %d", calls)
	}

	m.Delete("one")
	_, ok = m.Get("one")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestConcurrentMapConcurrentAccess(t *testing.T) {
	m := NewConcurrentMap[int, int]()

	var wg sync.WaitGroup
	for worker := range 16 {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for i := range 100 {
				key := worker*100 + i
				m.Set(key, i)
				m.Get(key)
				m.Len()
				if i%10 == 0 {
					m.Range(func(key int, value int) bool {
						return true
					})
				}
				if i%2 == 0 {
					m.Delete(key)
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestConcurrentMapRangeCanMutateMap(t *testing.T) {
	m := NewConcurrentMap[int, int]()
	m.Set(1, 1)

	m.Range(func(key int, value int) bool {
		m.Set(2, 2)
		m.Delete(key)
		return true
	})

	value, ok := m.Get(2)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != 2 {
		t.Fatalf("expected value 2, got %d", value)
	}
}
