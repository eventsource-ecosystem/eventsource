package eventsource_test

import (
	"reflect"
	"testing"

	"github.com/eventsource-ecosystem/eventsource"
)

type EntitySetName struct {
	eventsource.Model
	Name string
}

func TestJSONSerializer(t *testing.T) {
	event := EntitySetName{
		Model: eventsource.Model{
			ID:      "123",
			Version: 456,
		},
		Name: "blah",
	}

	serializer := eventsource.NewJSONSerializer(event)
	record, err := serializer.MarshalEvent(event)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	v, err := serializer.UnmarshalEvent(record)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	found, ok := v.(*EntitySetName)
	if !ok {
		t.Fatalf("got false; want true")
	}
	if got, want := found, &event; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestJSONSerializer_MarshalAll(t *testing.T) {
	event := EntitySetName{
		Model: eventsource.Model{
			ID:      "123",
			Version: 456,
		},
		Name: "blah",
	}

	serializer := eventsource.NewJSONSerializer(event)
	history, err := serializer.MarshalAll(event)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if history == nil {
		t.Fatalf("got nil; want not nil")
	}

	v, err := serializer.UnmarshalEvent(history[0])
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	found, ok := v.(*EntitySetName)
	if !ok {
		t.Fatalf("got false; want true")
	}
	if got, want := found, &event; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}
