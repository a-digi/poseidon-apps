package events

import (
	"sync"
)

type EventHandler func(data interface{})

type EventBus struct {
	listeners map[string][]EventHandler
	lock      sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(event string, handler EventHandler) {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	eb.listeners[event] = append(eb.listeners[event], handler)
}

func (eb *EventBus) Unsubscribe(event string, handler EventHandler) {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	handlers := eb.listeners[event]
	for i, h := range handlers {
		if &h == &handler {
			eb.listeners[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish löst ein Event aus und ruft alle zugehörigen Handler auf
func (eb *EventBus) Publish(event string, data interface{}) {
	eb.lock.RLock()
	handlers := eb.listeners[event]
	eb.lock.RUnlock()
	for _, handler := range handlers {
		go handler(data)
	}
}
