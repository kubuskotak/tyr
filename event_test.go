package tyr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func sender(c chan Event) {
	events := []EventType{
		ReadQuery,
		CreatedQuery,
		UpdatedQuery,
		DeletedQuery,
	}

	for n := 0; n < len(events); n++ {
		// send event to channel
		c <- Event{
			Type: events[n],
			Data: fmt.Sprintf("test sender - %d", n),
		}
	}
	close(c)
}

func TestNewEventHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handlers := NewEventHandler()
	handlers.Handle(ctx, ReadQuery, func(e Event) {
		require.Equal(t, ReadQuery, e.Type)
		t.Logf("Data %v :%s \n", e, string(ReadQuery))
	})
	handlers.Handle(ctx, CreatedQuery, func(e Event) {
		require.Equal(t, CreatedQuery, e.Type)
		t.Logf("Data %v :%s \n", e, string(CreatedQuery))
	})
	handlers.Handle(ctx, UpdatedQuery, func(e Event) {
		require.Equal(t, UpdatedQuery, e.Type)
		t.Logf("Data %v :%s \n", e, string(UpdatedQuery))
	})
	handlers.Handle(ctx, DeletedQuery, func(e Event) {
		require.Equal(t, DeletedQuery, e.Type)
		t.Logf("Data %v :%s \n", e, string(DeletedQuery))
	})

	events := make(chan Event)
	go sender(events)
	for e := range events {
		handlers.Notify(ctx, e)
	}

	cancel()
}
