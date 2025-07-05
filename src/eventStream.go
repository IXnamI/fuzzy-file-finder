package fff

import "sync"

type EventStream struct {
	events   chan Event
	capacity int
	mu       sync.Mutex
}

type Event struct {
	Type EventType
	Char rune
}

type EventType int

const (
	Rune EventType = iota
	EnterKey
	CtrlC
	BackSpace
)

func NewEventStream(capacity int) *EventStream {
	return &EventStream{
		events:   make(chan Event, capacity),
		capacity: capacity,
	}
}

func NewEvent(eventType EventType, char rune) Event {
	return Event{Type: eventType, Char: char}
}

func (es *EventStream) Drain(handler func(Event)) {
	es.mu.Lock()
	defer es.mu.Unlock()
	for {
		select {
		case event := <-es.events:
			handler(event)
		default:
			return
		}
	}
}

func (es *EventStream) AddEvent(newEvent Event) {
	es.events <- newEvent
}
