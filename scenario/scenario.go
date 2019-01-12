package scenario

import (
	"context"
	"reflect"
	"strings"

	"github.com/eventsource-ecosystem/eventsource"
)

// TestingT is a wrapper for *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

// CommandHandlerAggregate implements both Aggregate and CommandHandler
type CommandHandlerAggregate interface {
	eventsource.CommandHandler
	eventsource.Aggregate
}

// Builder captures the data used to execute a test scenario
type Builder struct {
	t         TestingT
	aggregate CommandHandlerAggregate
	given     []eventsource.Event
	command   eventsource.Command
}

func (b *Builder) clone() *Builder {
	return &Builder{
		t:         b.t,
		aggregate: b.aggregate,
		given:     b.given,
		command:   b.command,
	}
}

// Given allows an initial set of events to be provided; may be called multiple times
func (b *Builder) Given(given ...eventsource.Event) *Builder {
	dupe := b.clone()
	dupe.given = append(dupe.given, given...)
	return dupe
}

// When provides the command to test
func (b *Builder) When(command eventsource.Command) *Builder {
	dupe := b.clone()
	dupe.command = command
	return dupe
}

func (b *Builder) apply() ([]eventsource.Event, error) {
	t := reflect.TypeOf(b.aggregate)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	aggregate := reflect.New(t).Interface().(CommandHandlerAggregate)

	// given
	for _, e := range b.given {
		if got := aggregate.On(e); got != nil {
			b.t.Errorf("got %v; want nil", got)
		}
	}

	// when
	ctx := context.Background()
	return aggregate.Apply(ctx, b.command)
}

// deepEquals (unlike reflect.DeepEqual) only performs a deep equality check on non-zero fields
func deepEquals(t TestingT, expected, actual interface{}, path ...string) bool {
	te := reflect.TypeOf(expected)
	ta := reflect.TypeOf(actual)
	if got, want := ta, te; !reflect.DeepEqual(got, want) {
		return false
	}

	if te.Kind() == reflect.Ptr {
		te = te.Elem()
	}

	ve := reflect.ValueOf(expected)
	if ve.Kind() == reflect.Ptr {
		ve = ve.Elem()
	}

	va := reflect.ValueOf(actual)
	if va.Kind() == reflect.Ptr {
		va = va.Elem()
	}

	for i := 0; i < te.NumField(); i++ {
		fieldName := te.Field(i).Name
		fieldType := te.Field(i).Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		fe := ve.Field(i)
		fa := va.Field(i)

		if !fe.CanInterface() || !fa.CanInterface() {
			continue
		}
		if zero := reflect.Zero(fieldType).Interface(); reflect.DeepEqual(zero, fe.Interface()) {
			continue
		}

		if fieldType.Kind() == reflect.Struct {
			if got, want := fa.Interface(), fe.Interface(); !deepEquals(t, got, want, append(path, fieldName)...) {
				return false
			}
			continue
		}

		if got, want := fa.Interface(), fe.Interface(); !reflect.DeepEqual(got, want) {
			t.Errorf("%v.%v: got %v; want %v", strings.Join(path, "."), fieldName, got, want)
			return false
		}
	}

	return true
}

// Then check that the command returns the following events.  Only non-zero valued
// fields will be checked.  If no non-zeroed values are present, then only the
// event type will be checked
func (b *Builder) Then(expected ...eventsource.Event) {
	actual, err := b.apply()
	if err != nil {
		b.t.Errorf("got %v; want nil", err)
	}

	// then
	if got, want := len(actual), len(expected); got != want {
		b.t.Errorf("got %v; want %v", got, want)
		return
	}

	for index, e := range expected {
		a := actual[index]
		deepEquals(b.t, e, a)
	}
}

// ThenError verifies that the error returned by the command matches
// the function expectation
func (b *Builder) ThenError(matches func(err error) bool) {
	_, err := b.apply()
	if got := matches(err); !got {
		b.t.Errorf("got false; want true")
	}
}

// New constructs a new scenario
func New(t TestingT, prototype CommandHandlerAggregate) *Builder {
	return &Builder{
		t:         t,
		aggregate: prototype,
	}
}
