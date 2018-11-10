package eventsource_test

import (
	"testing"

	"github.com/eventsource-ecosystem/eventsource"
)

func TestCommandModel_AggregateID(t *testing.T) {
	m := eventsource.CommandModel{ID: "abc"}
	if got, want := m.ID, m.AggregateID(); got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}
