package eventsource

import (
	"reflect"
	"testing"
)

type Entity struct {
}

func (item *Entity) On(event Event) error {
	return nil
}

func TestNew(t *testing.T) {
	repository := New(&Entity{})
	aggregate := repository.newAggregate()
	if got := aggregate; got == nil {
		t.Fatalf("got nil; want not nil")
	}

	want := &Entity{}
	if got := aggregate; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want not %v", got, want)
	}
}
