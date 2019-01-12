package scenario_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eventsource-ecosystem/eventsource"
	"github.com/eventsource-ecosystem/eventsource/scenario"
)

//Order is an example of state generated from left fold of events
type Order struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	State     string
}

//OrderCreated event used a marker of order created
type OrderCreated struct {
	eventsource.Model
}

//OrderShipped event used a marker of order shipped
type OrderShipped struct {
	eventsource.Model
}

//On implements Aggregate interface
func (item *Order) On(event eventsource.Event) error {
	switch v := event.(type) {
	case *OrderCreated:
		item.State = "created"

	case *OrderShipped:
		item.State = "shipped"

	default:
		return fmt.Errorf("unable to handle event, %v", v)
	}

	item.Version = event.EventVersion()
	item.ID = event.AggregateID()
	item.UpdatedAt = event.EventAt()

	return nil
}

//CreateOrder command
type CreateOrder struct {
	eventsource.CommandModel
}

//ShipOrder command
type ShipOrder struct {
	eventsource.CommandModel
}

func (item *Order) Apply(ctx context.Context, command eventsource.Command) ([]eventsource.Event, error) {
	switch v := command.(type) {
	case *CreateOrder:
		orderCreated := &OrderCreated{
			Model: eventsource.Model{ID: command.AggregateID(), Version: item.Version + 1, At: time.Now()},
		}
		return []eventsource.Event{orderCreated}, nil

	case *ShipOrder:
		if item.State != "created" {
			return nil, fmt.Errorf("order, %v, has already shipped", command.AggregateID())
		}
		orderShipped := &OrderShipped{
			Model: eventsource.Model{ID: command.AggregateID(), Version: item.Version + 1, At: time.Now()},
		}
		return []eventsource.Event{orderShipped}, nil

	default:
		return nil, fmt.Errorf("unhandled command, %v", v)
	}
}

func TestSimpleScenario(t *testing.T) {
	var (
		prototype = &Order{}
		command   = &CreateOrder{}
		event     = &OrderCreated{}
	)

	scenario.Test(t, prototype).
		Given().
		When(command).
		Then(event)
}

type Errors struct {
	Messages []string
}

func (e *Errors) Errorf(format string, args ...interface{}) {
	e.Messages = append(e.Messages, fmt.Sprintf(format, args...))
}

func TestFieldError(t *testing.T) {
	const id = "abc"

	t.Run("ok", func(t *testing.T) {
		errs := &Errors{}
		scenario.Test(errs, &Order{}).
			When(
				&CreateOrder{CommandModel: eventsource.CommandModel{ID: id}},
			).
			Then(
				&OrderCreated{Model: eventsource.Model{ID: id + "junk"}},
			)

		if got, want := len(errs.Messages), 1; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
		if got, want := errs.Messages[0], "junk"; !strings.Contains(got, want) {
			t.Fatalf("expected %v to contain %v", got, want)
		}
	})

	t.Run("fail - expect order shipped", func(t *testing.T) {
		errs := &Errors{}
		scenario.Test(errs, &Order{}).
			When(
				&CreateOrder{CommandModel: eventsource.CommandModel{ID: id}},
			).
			Then(
				&OrderShipped{Model: eventsource.Model{ID: id, Version: 1}},
			)

		if got, want := len(errs.Messages), 1; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
		if got, want := errs.Messages[0], "got OrderCreated; want OrderShipped"; !strings.Contains(got, want) {
			t.Fatalf("expected %v to contain %v", got, want)
		}
	})

	t.Run("pass", func(t *testing.T) {
		scenario.Test(t, &Order{}).
			When(
				&CreateOrder{CommandModel: eventsource.CommandModel{ID: id}},
			).
			Then(
				&OrderCreated{Model: eventsource.Model{ID: id, Version: 1}},
			)
	})
}

func TestDeepEquals(t *testing.T) {
	if got, want := reflect.DeepEqual([]uint8(nil), []byte(nil)), true; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}
