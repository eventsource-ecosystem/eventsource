package eventsource

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/xerrors"
)

// Record provides the serialized representation of the event
type Record struct {
	// Version contains the version associated with the serialized event
	Version int

	// Data contains the event in serialized form
	Data []byte
}

// History represents
type History []Record

// Len implements sort.Interface
func (h History) Len() int {
	return len(h)
}

// Swap implements sort.Interface
func (h History) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Less implements sort.Interface
func (h History) Less(i, j int) bool {
	return h[i].Version < h[j].Version
}

// Store provides an abstraction for the Repository to save data
type Store interface {
	// Save the provided serialized records to the store
	Save(ctx context.Context, aggregateID string, records ...Record) error

	// Load the history of events up to the version specified.
	// When toVersion is 0, all events will be loaded.
	// To start at the beginning, fromVersion should be set to 0
	Load(ctx context.Context, aggregateID string, fromVersion, toVersion int) (History, error)
}

// Deprecated - use AggregateSaver  instead
//
// StoreAggregate provides an alternative to Store that allows the aggregate to
// be saved at the same time as the events.  When the Store implement StoreAggregate,
// the SaveAggregate will always be called instead of Store
type StoreAggregate interface {
	// Save the provided serialized records to the store
	SaveAggregate(ctx context.Context, aggregateID string, aggregate Aggregate, records ...Record) error
}

// SaveAggregateInput includes additional fields that simplify the saving of Aggregates
type SaveAggregateInput struct {
	AggregateID string    // AggregateID
	Aggregate   Aggregate // Aggregate (with all the events applied)
	Events      []Event   // Events that were applied
	Records     []Record  // Records to be persisted e.g. the serialized events
}

// AggregateSaver provides a custom interface to save aggregates
type AggregateSaver interface {
	SaveAggregate(ctx context.Context, input SaveAggregateInput) error
}

// memoryStore provides an in-memory implementation of Store
type memoryStore struct {
	mux        *sync.Mutex
	eventsByID map[string]History
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		mux:        &sync.Mutex{},
		eventsByID: map[string]History{},
	}
}

func (m *memoryStore) Save(ctx context.Context, aggregateID string, records ...Record) error {
	if _, ok := m.eventsByID[aggregateID]; !ok {
		m.eventsByID[aggregateID] = History{}
	}

	history := append(m.eventsByID[aggregateID], records...)
	sort.Sort(history)
	m.eventsByID[aggregateID] = history

	return nil
}

func (m *memoryStore) Load(ctx context.Context, aggregateID string, fromVersion, toVersion int) (History, error) {
	all, ok := m.eventsByID[aggregateID]
	if !ok {
		return nil, xerrors.Errorf("no aggregate found with id, %v: %w", aggregateID, errAggregateNotFound)
	}

	history := make(History, 0, len(all))
	if len(all) > 0 {
		for _, record := range all {
			if v := record.Version; v >= fromVersion && (toVersion == 0 || v <= toVersion) {
				history = append(history, record)
			}
		}
	}

	return all, nil
}
