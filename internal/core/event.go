package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type EventHandler func(ctx context.Context, topic string, event *Event)

type Subscription struct {
	ID   string
	topic string
}

type EventBus interface {
	Publish(topic string, event *Event)
	Subscribe(topic string, handler EventHandler) Subscription
	SubscribeAny(handler EventHandler) Subscription
	Unsubscribe(sub Subscription)
	Close()
}

type InMemoryEventBus struct {
	mu          sync.RWMutex
	handlers    map[string][]HandlerEntry
	anyHandlers []HandlerEntry
	closed      bool
	nextID      int
}

type HandlerEntry struct {
	ID      string
	Handler EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]HandlerEntry),
	}
}

func (b *InMemoryEventBus) Publish(topic string, event *Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if handlers, ok := b.handlers[topic]; ok {
		for _, entry := range handlers {
			go entry.Handler(context.Background(), topic, event)
		}
	}
	for _, entry := range b.anyHandlers {
		go entry.Handler(context.Background(), topic, event)
	}
}

func (b *InMemoryEventBus) Subscribe(topic string, handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := fmt.Sprintf("sub-%d", b.nextID)
	b.handlers[topic] = append(b.handlers[topic], HandlerEntry{ID: id, Handler: handler})
	return Subscription{ID: id, topic: topic}
}

func (b *InMemoryEventBus) SubscribeAny(handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := fmt.Sprintf("sub-%d", b.nextID)
	b.anyHandlers = append(b.anyHandlers, HandlerEntry{ID: id, Handler: handler})
	return Subscription{ID: id}
}

func (b *InMemoryEventBus) Unsubscribe(sub Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sub.topic != "" {
		handlers := b.handlers[sub.topic]
		for i, entry := range handlers {
			if entry.ID == sub.ID {
				b.handlers[sub.topic] = append(handlers[:i], handlers[i+1:]...)
				return
			}
		}
	}
	for i, entry := range b.anyHandlers {
		if entry.ID == sub.ID {
			b.anyHandlers = append(b.anyHandlers[:i], b.anyHandlers[i+1:]...)
			return
		}
	}
}

func (b *InMemoryEventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.handlers = nil
	b.anyHandlers = nil
}
