package eventsource

import "golang.org/x/xerrors"

type errorType string

func (e errorType) Error() string {
	return string(e)
}

const (
	//AggregateNil      = "AggregateNil"
	//DuplicateID       = "DuplicateID"
	//DuplicateVersion  = "DuplicateVersion"
	//DuplicateAt       = "DuplicateAt"
	//DuplicateType     = "DuplicateType"
	//InvalidID         = "InvalidID"
	//InvalidAt         = "InvalidAt"
	//InvalidVersion    = "InvalidVersion"

	// InvalidEncoding is returned when the Serializer cannot marshal the event
	errInvalidEncoding errorType = "InvalidEncoding"

	// UnboundEventType when the Serializer cannot unmarshal the serialized event
	errUnboundEventType errorType = "UnboundEventType"

	// AggregateNotFound will be returned when attempting to Load an aggregateID
	// that does not exist in the Store
	errAggregateNotFound errorType = "AggregateNotFound"

	// UnhandledEvent occurs when the Aggregate is unable to handle an event and returns
	// a non-nil err
	errUnhandledEvent errorType = "UnhandledEvent"
)

// IsNotFound returns true if the error was AggregateNotFound
func IsNotFoundError(err error) bool {
	return xerrors.Is(err, errAggregateNotFound)
}
