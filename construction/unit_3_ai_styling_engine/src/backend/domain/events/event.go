package events

// DomainEvent is the base interface all in-process domain events must implement.
type DomainEvent interface {
	EventType() string
}

// EventHandler is a function that handles a domain event.
type EventHandler func(event DomainEvent)

// EventDispatcher dispatches domain events to registered handlers in-process.
type EventDispatcher interface {
	Register(eventType string, handler EventHandler)
	Dispatch(event DomainEvent)
}
