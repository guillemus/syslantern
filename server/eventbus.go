package server

import (
	"context"
	"sync"
)

// EventBus is a custom fire and forget implementation of an event bus.
// Emitted events are meant to be interepreted as state invalidations.
type EventBus[T any] struct {
	mu   sync.RWMutex
	subs []chan T
}

func NewEventBus[T any]() *EventBus[T] {
	return &EventBus[T]{
		subs: []chan T{},
	}
}

func (b *EventBus[T]) Subscribe(ctx context.Context) <-chan T {
	ch := make(chan T, 16)

	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		// remove sub from b.subs
		for i, sub := range b.subs {
			if sub == ch {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				close(ch)
				break
			}
		}
	}()

	return ch
}

func (b *EventBus[T]) Emit(msg T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subs {
		select {
		case sub <- msg:
		default:
		}
	}
}
