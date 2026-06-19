package server

import (
	"context"
	"sync"
)

type EventBus[T any] struct {
	mu   sync.RWMutex
	subs []*func(T)
}

func NewEventBus[T any]() *EventBus[T] {
	return &EventBus[T]{
		subs: []*func(T){},
	}
}

func (b *EventBus[T]) Subscribe(ctx context.Context, fn func(T)) context.CancelFunc {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, cancelCtx := context.WithCancel(ctx)

	fnPtr := &fn
	b.subs = append(b.subs, fnPtr)
	cancel := func() {
		defer cancelCtx()

		b.mu.Lock()
		defer b.mu.Unlock()

		// Remove the subscriber from the list
		for i, sub := range b.subs {
			if sub == fnPtr {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				break
			}
		}
	}
	return cancel
}

func (b *EventBus[T]) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

func (b *EventBus[T]) Emit(ctx context.Context, msg T) {
	// Subs might be modified while we are iterating over them,
	// so we need to copy them first.
	b.mu.RLock()
	subs := make([]*func(T), len(b.subs))
	copy(subs, b.subs)
	b.mu.RUnlock()

	for _, sub := range subs {
		go (*sub)(msg)
	}
}
