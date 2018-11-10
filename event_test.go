package eventsource_test

import (
	"testing"
	"time"

	"github.com/eventsource-ecosystem/eventsource"
)

func TestEvent(t *testing.T) {
	m := eventsource.Model{
		ID:      "abc",
		Version: 123,
		At:      time.Now(),
	}

	if got, want := m.AggregateID(), m.ID; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
	if got, want := m.EventVersion(), m.Version; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
	if got, want := m.EventAt(), m.At; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}
