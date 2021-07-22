package tyr

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
)

type EventType string

const (
	ReadQuery    EventType = "READ_QUERY"
	CreatedQuery EventType = "CREATE_QUERY"
	UpdatedQuery EventType = "UPDATE_QUERY"
	DeletedQuery EventType = "DELETE_QUERY"
)

type EventFunc func(Event)

type EventMap map[EventType][]EventFunc

type Event struct {
	Type EventType
	Data interface{}
}

type EventHandler struct {
	Event    Event
	EventMap EventMap
}

func NewEventHandler() *EventHandler {
	eventMap := make(EventMap)

	return &EventHandler{
		EventMap: eventMap,
	}
}

// Notify function to invoke event handle
func (h *EventHandler) Notify(ctx context.Context, event Event) {
	span, _ := opentracing.StartSpanFromContext(ctx, "tyr.Event.Notify")
	defer span.Finish()
	h.Event = event
	h.Dispatcher(ctx)
}

// Handle register the handler function to handle an event type
func (h *EventHandler) Handle(ctx context.Context, e EventType, f EventFunc) {
	span, _ := opentracing.StartSpanFromContext(ctx, "tyr.Event.Handle")
	defer span.Finish()
	h.EventMap[e] = append(h.EventMap[e], f)
}

// Dispatcher the handler function to invoke event
func (h *EventHandler) Dispatcher(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "tyr.Event.Dispatcher")
	defer span.Finish()

	if handlers, ok := h.EventMap[h.Event.Type]; ok {
		fmt.Printf("event received: %v \n", h.Event)
		for _, fn := range handlers {
			go fn(h.Event)
			time.Sleep(5 * time.Millisecond) // for sync hook event
		}
	}
}
