package eventsource_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/eventsource-ecosystem/eventsource"
)

func TestNewError(t *testing.T) {
	err := eventsource.NewError(io.EOF, "code", "hello %v", "world")
	if err == nil {
		t.Fatalf("got nil; want not nil")
	}

	v, ok := err.(eventsource.Error)
	if !ok {
		t.Fatalf("got false; want true")
	}
	if got, want := v.Cause(), io.EOF; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
	if got, want := v.Code(), "code"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
	if got, want := v.Message(), "hello world"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
	if got, want := v.Error(), "[code] hello world - EOF"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	s, ok := err.(fmt.Stringer)
	if !ok {
		t.Fatalf("got false; want true")
	}
	if got, want := s.String(), v.Error(); got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestIsNotFound(t *testing.T) {
	testCases := map[string]struct {
		Err        error
		IsNotFound bool
	}{
		"nil": {
			Err:        nil,
			IsNotFound: false,
		},
		"eventsource.Error": {
			Err:        eventsource.NewError(nil, eventsource.ErrAggregateNotFound, "not found"),
			IsNotFound: true,
		},
		"nested eventsource.Error": {
			Err: eventsource.NewError(
				eventsource.NewError(nil, eventsource.ErrAggregateNotFound, "not found"),
				eventsource.ErrUnboundEventType,
				"not found",
			),
			IsNotFound: true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			if got, want := eventsource.IsNotFound(tc.Err), tc.IsNotFound; got != want {
				t.Fatalf("got %v; want %v", got, want)
			}
		})
	}
}

func TestErrHasCode(t *testing.T) {
	code := "code"

	testCases := map[string]struct {
		Err        error
		ErrHasCode bool
	}{
		"simple": {
			Err:        eventsource.NewError(nil, code, "blah"),
			ErrHasCode: true,
		},
		"nope": {
			Err:        errors.New("blah"),
			ErrHasCode: false,
		},
		"nested": {
			Err:        eventsource.NewError(eventsource.NewError(nil, code, "blah"), "blah", "blah"),
			ErrHasCode: true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			if got, want := eventsource.ErrHasCode(tc.Err, code), tc.ErrHasCode; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v; want %v", got, want)
			}
		})
	}
}
