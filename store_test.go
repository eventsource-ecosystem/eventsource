package eventsource_test

import (
	"sort"
	"testing"

	"github.com/eventsource-ecosystem/eventsource"
)

func TestHistory_Swap(t *testing.T) {
	history := eventsource.History{
		{Version: 3},
		{Version: 1},
		{Version: 2},
	}

	sort.Sort(history)
	if got, want := history[0].Version, 1; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
	if got, want := history[1].Version, 2; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
	if got, want := history[2].Version, 3; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}
