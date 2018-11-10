package eventsource_test

import (
	"context"
	"errors"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/eventsource-ecosystem/eventsource"
)

type Entity struct {
	Version   int
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EntityCreated struct {
	eventsource.Model
}

type EntityNameSet struct {
	eventsource.Model
	Name string
}

func (item *Entity) On(event eventsource.Event) error {
	switch v := event.(type) {
	case *EntityCreated:
		item.Version = v.Model.Version
		item.ID = v.Model.ID
		item.CreatedAt = v.Model.At
		item.UpdatedAt = v.Model.At

	case *EntityNameSet:
		item.Version = v.Model.Version
		item.Name = v.Name
		item.UpdatedAt = v.Model.At

	default:
		return errors.New(eventsource.ErrUnhandledEvent)
	}

	return nil
}

type CreateEntity struct {
	eventsource.CommandModel
}

type Nop struct {
	eventsource.CommandModel
}

func (item *Entity) Apply(ctx context.Context, command eventsource.Command) ([]eventsource.Event, error) {
	switch command.(type) {
	case *CreateEntity:
		return []eventsource.Event{&EntityCreated{
			Model: eventsource.Model{
				ID:      command.AggregateID(),
				Version: item.Version + 1,
				At:      time.Now(),
			},
		}}, nil

	case *Nop:
		return []eventsource.Event{}, nil

	default:
		return []eventsource.Event{}, nil
	}
}

func TestNew(t *testing.T) {
	repository := eventsource.New(&Entity{})
	aggregate := repository.New()
	if got := aggregate; got == nil {
		t.Fatalf("got nil; want not nil")
	}

	want := &Entity{}
	if got := aggregate; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want not %v", got, want)
	}
}

func TestRepository_Load_NotFound(t *testing.T) {
	ctx := context.Background()
	repository := eventsource.New(&Entity{},
		eventsource.WithDebug(ioutil.Discard),
	)

	_, err := repository.Load(ctx, "does-not-exist")
	if err == nil {
		t.Fatalf("got nil; want not nil")
	}
	if !eventsource.IsNotFound(err) {
		t.Fatalf("got false; want true")
	}
}

func TestRegistry(t *testing.T) {
	ctx := context.Background()
	id := "123"
	name := "Jones"
	serializer := eventsource.NewJSONSerializer(
		EntityCreated{},
		EntityNameSet{},
	)

	t.Run("simple", func(t *testing.T) {
		repository := eventsource.New(&Entity{},
			eventsource.WithSerializer(serializer),
			eventsource.WithDebug(ioutil.Discard),
		)

		// Test - Add an event to the store and verify we can recreate the object

		err := repository.Save(ctx,
			&EntityCreated{
				Model: eventsource.Model{ID: id, Version: 0, At: time.Unix(3, 0)},
			},
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 1, At: time.Unix(4, 0)},
				Name:  name,
			},
		)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		v, err := repository.Load(ctx, id)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		org, ok := v.(*Entity)
		if !ok {
			t.Fatalf("got false; want true")
		}
		if got, want := org.ID, id; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
		if got, want := org.Name, name; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}

		// Test - Update the org name and verify that the change is reflected in the loaded result

		updated := "Sarah"
		err = repository.Save(ctx, &EntityNameSet{
			Model: eventsource.Model{ID: id, Version: 2},
			Name:  updated,
		})
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		v, err = repository.Load(ctx, id)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		org, ok = v.(*Entity)
		if !ok {
			t.Fatalf("got false; want true")
		}
		if got, want := org.ID, id; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
		if got, want := org.Name, updated; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("with pointer prototype", func(t *testing.T) {
		registry := eventsource.New(&Entity{},
			eventsource.WithSerializer(serializer),
		)

		err := registry.Save(ctx,
			&EntityCreated{
				Model: eventsource.Model{ID: id, Version: 0, At: time.Unix(3, 0)},
			},
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 1, At: time.Unix(4, 0)},
				Name:  name,
			},
		)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		v, err := registry.Load(ctx, id)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if got, want := v.(*Entity).Name, name; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("with pointer bind", func(t *testing.T) {
		registry := eventsource.New(&Entity{},
			eventsource.WithSerializer(serializer),
		)

		err := registry.Save(ctx,
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 0},
				Name:  name,
			},
		)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		v, err := registry.Load(ctx, id)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if got, want := v.(*Entity).Name, name; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})
}

func TestAt(t *testing.T) {
	ctx := context.Background()
	id := "123"

	registry := eventsource.New(&Entity{},
		eventsource.WithSerializer(eventsource.NewJSONSerializer(EntityCreated{})),
	)

	err := registry.Save(ctx,
		&EntityCreated{
			Model: eventsource.Model{ID: id, Version: 1, At: time.Now()},
		},
	)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	v, err := registry.Load(ctx, id)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	org := v.(*Entity)
	if org.CreatedAt.IsZero() {
		t.Fatalf("got %v; want zero", org.CreatedAt)
	}
	if org.UpdatedAt.IsZero() {
		t.Fatalf("got %v; want zero", org.UpdatedAt)
	}
}

func TestRepository_SaveNoEvents(t *testing.T) {
	repository := eventsource.New(&Entity{})
	err := repository.Save(context.Background())
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
}

func TestWithObservers(t *testing.T) {
	var captured []eventsource.Event
	observer := func(event eventsource.Event) {
		captured = append(captured, event)
	}

	repository := eventsource.New(&Entity{},
		eventsource.WithSerializer(
			eventsource.NewJSONSerializer(
				EntityCreated{},
				EntityNameSet{},
			),
		),
		eventsource.WithDebug(ioutil.Discard),
		eventsource.WithObservers(observer),
	)

	ctx := context.Background()

	// When I dispatch command
	err := repository.Dispatch(ctx, &CreateEntity{
		CommandModel: eventsource.CommandModel{ID: "abc"},
	})

	// Then I expect event to be captured
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got, want := len(captured), 1; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	_, ok := captured[0].(*EntityCreated)
	if !ok {
		t.Fatalf("got false; want true")
	}
}

func TestApply(t *testing.T) {
	repo := eventsource.New(&Entity{},
		eventsource.WithSerializer(
			eventsource.NewJSONSerializer(
				EntityCreated{},
			),
		),
	)

	cmd := &CreateEntity{CommandModel: eventsource.CommandModel{ID: "123"}}

	// When
	version, err := repo.Apply(context.Background(), cmd)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got, want := version, 1; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// And
	version, err = repo.Apply(context.Background(), cmd)
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if got, want := version, 2; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestApplyNopCommand(t *testing.T) {
	t.Run("Version still returned when command generates no events", func(t *testing.T) {
		repo := eventsource.New(&Entity{},
			eventsource.WithSerializer(
				eventsource.NewJSONSerializer(
					EntityCreated{},
				),
			),
		)

		cmd := &Nop{
			CommandModel: eventsource.CommandModel{ID: "abc"},
		}
		version, err := repo.Apply(context.Background(), cmd)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if got, want := version, 0; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})
}
