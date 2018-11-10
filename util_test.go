package eventsource_test

import (
	"testing"

	"github.com/eventsource-ecosystem/eventsource"
)

type Custom struct {
	eventsource.Model
}

func (c Custom) EventType() string {
	return "blah"
}

func TestEventType(t *testing.T) {
	m := Custom{}
	eventType, _ := eventsource.EventType(m)
	if got, want := eventType, "blah"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}
