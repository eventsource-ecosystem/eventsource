package eventsource

import "testing"

func Test_errorType_Error(t *testing.T) {
	if got, want := errAggregateNotFound.Error(), string(errAggregateNotFound); got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}
