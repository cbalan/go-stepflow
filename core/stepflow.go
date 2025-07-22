package core

import (
	"context"
	"fmt"
)

// StepFlow represents an executable workflow. It applies transitions to move from one state to another.
type StepFlow interface {
	// Apply executes the workflow from the given state, returning the new state and any error.
	Apply(ctx context.Context, state []string) ([]string, error)

	// IsCompleted checks if the workflow has reached its completion state.
	IsCompleted(state []string) bool
}

// stepFlowImpl implements the StepFlow interface and manages the execution of a workflow.
type stepFlowImpl struct {
	item           StepFlowItem
	transitionsMap map[string][]Transition
	startState     []string
	completedState []string
}

// NewStepFlow creates a new executable workflow using the provided step flow item as a root item.
func NewStepFlow(item StepFlowItem) (StepFlow, error) {
	itemScope, transitions, err := item.Transitions(nil)
	if err != nil {
		return nil, err
	}

	transitionsMap := make(map[string][]Transition)
	for _, t := range transitions {
		source := eventString(t.Source())
		transitionsMap[source] = append(transitionsMap[source], t)
	}

	startState := []string{eventString(StartCommand(itemScope))}
	completedState := []string{eventString(CompletedEvent(itemScope))}

	return &stepFlowImpl{item: item, transitionsMap: transitionsMap, startState: startState, completedState: completedState}, nil
}

// ApplyOneMaxIterations limits the maximum number of state transitions in a single Apply call
// to prevent infinite loops in workflows that don't terminate properly.
const ApplyOneMaxIterations = 100

// Apply executes the workflow starting from the given state (or the default start state if nil).
// It repeatedly applies transitions until an error occurs, an exclusive transition is encountered,
// or the maximum number of iterations is reached.
func (sf *stepFlowImpl) Apply(ctx context.Context, oldState []string) ([]string, error) {
	newState := withDefaultValue(oldState, sf.startState)
	var isExclusive bool
	var err error

	for range ApplyOneMaxIterations {
		newState, isExclusive, err = sf.applyOne(ctx, newState)
		if err != nil || isExclusive {
			break
		}
	}

	return newState, err
}

// applyOne performs a single transition from the current state to the next state.
// It returns the new state, whether the transition is exclusive, and any error that occurred.
func (sf *stepFlowImpl) applyOne(ctx context.Context, oldState []string) ([]string, bool, error) {
	if sf.IsCompleted(oldState) {
		return oldState, true, nil
	}

	for _, lastEvent := range oldState {
		for _, t := range sf.transitionsMap[lastEvent] {
			isExclusive := t.IsExclusive()
			newState, err := t.Destination(ctx)
			return eventsString(newState), isExclusive, err
		}
	}

	return nil, true, fmt.Errorf("unhandled state %s", oldState)
}

// withDefaultValue returns the default value if the given value is nil, otherwise returns the value.
func withDefaultValue(value []string, defaultValue []string) []string {
	if value == nil {
		return defaultValue
	}

	return value
}

// IsCompleted checks if the workflow has reached its completion state.
func (sf *stepFlowImpl) IsCompleted(state []string) bool {
	if len(state) != 1 {
		return false
	}

	return state[0] == sf.completedState[0]
}

// Scope represents a named context in which events occur. Scopes can be nested to allow hierarchical structures.
type Scope interface {
	// Name returns the fully qualified name of the scope.
	Name() string

	// Parent returns the parent scope, or nil if this is a root scope.
	Parent() Scope
}

// scopeImpl is the concrete implementation of the Scope interface
type scopeImpl struct {
	name   string
	parent Scope
}

// NewScope creates a new root scope with the given name
func NewScope(name string) Scope {
	return &scopeImpl{name: name}
}

// WithParent creates a new scope with the given parent, combining their names with a slash
// If parent is nil, returns the original scope unchanged
func WithParent(scope Scope, parent Scope) Scope {
	if parent == nil {
		return scope
	}

	return &scopeImpl{name: parent.Name() + "/" + scope.Name(), parent: parent}
}

// Name returns the fully qualified name of the scope.
func (s *scopeImpl) Name() string {
	return s.name
}

// Parent returns the parent scope, or nil if this is a root scope.
func (s *scopeImpl) Parent() Scope {
	return s.parent
}

// Event represents a specific point in the workflow execution, with a name and associated scope.
type Event interface {
	// Name returns the name of the event.
	Name() string
	// Scope returns the scope in which this event occurs.
	Scope() Scope
}

// eventImpl is the concrete implementation of the Event interface.
type eventImpl struct {
	name  string
	scope Scope
}

// NewEvent creates a new event with the given name and scope.
func NewEvent(name string, scope Scope) Event {
	return &eventImpl{name: name, scope: scope}
}

// Name returns the name of the event.
func (e eventImpl) Name() string {
	return e.name
}

// Scope returns the scope in which this event occurs.
func (e eventImpl) Scope() Scope {
	return e.scope
}

// eventString converts an event to a string representation for storage in the state.
func eventString(event Event) string {
	return event.Name() + ":" + event.Scope().Name()
}

// eventsString converts a slice of events to their string representations.
func eventsString(events []Event) []string {
	var result []string
	for _, event := range events {
		result = append(result, eventString(event))
	}

	return result
}

// StartCommand creates a "start" event for the given scope.
func StartCommand(scope Scope) Event {
	return NewEvent("start", scope)
}

// CompletedEvent creates a "completed" event for the given scope.
func CompletedEvent(scope Scope) Event {
	return NewEvent("completed", scope)
}

// StepFlowItem is the base interface for all workflow components.
// Each item defines its transitions within a parent scope.
type StepFlowItem interface {
	// Transitions returns the scope for this item, all transitions it defines,
	// and any error that occurred during creation.
	Transitions(parent Scope) (Scope, []Transition, error)
}

// PossibleDestination represents a potential destination event for a transition,
// along with a reason explaining why this destination might be chosen.
type PossibleDestination interface {
	// Event returns the potential destination event.
	Event() Event

	// Reason returns a human-readable explanation for this destination.
	Reason() string
}

// eventAndReason is the concrete implementation of the PossibleDestination interface.
type eventAndReason struct {
	event  Event
	reason string
}

// NewReason creates a new PossibleDestination with the given event and reason.
func NewReason(event Event, reason string) PossibleDestination {
	return &eventAndReason{event: event, reason: reason}
}

// Event returns the potential destination event.
func (n eventAndReason) Event() Event {
	return n.event
}

// Reason returns a human-readable explanation for this destination.
func (n eventAndReason) Reason() string {
	return n.reason
}

// Transition defines movement from one event to another in the workflow.
type Transition interface {
	// Source returns the event that triggers this transition.
	Source() Event

	// Destination evaluates the transition and returns the resulting events.
	Destination(context.Context) ([]Event, error)

	// IsExclusive indicates whether this transition can be applied
	// next to other transitions in the same StepFlow.Apply cycle.
	IsExclusive() bool

	// PossibleDestinations returns all potential destination events with their reasons.
	PossibleDestinations() []PossibleDestination
}

// staticTransition is a transition that always moves to predetermined destination events.
type staticTransition struct {
	source      Event
	destination []Event
}

// NewStaticTransition creates a new static transition from the source event to one or more destination events.
func NewStaticTransition(source Event, destination ...Event) Transition {
	return &staticTransition{source: source, destination: destination}
}

// Source implements the Transition interface.
func (t *staticTransition) Source() Event {
	return t.source
}

// Destination implements the Transition interface.
func (t *staticTransition) Destination(_ context.Context) ([]Event, error) {
	return t.destination, nil
}

// IsExclusive implements the Transition interface.
// Static transitions are not exclusive and allow other transitions to be applied.
func (t *staticTransition) IsExclusive() bool {
	return false
}

// PossibleDestinations implements the Transition interface.
func (t *staticTransition) PossibleDestinations() []PossibleDestination {
	var result []PossibleDestination
	for _, destination := range t.destination {
		result = append(result, NewReason(destination, "static"))
	}
	return result
}

// dynamicTransition is a transition that uses a function to determine the destination events at runtime.
type dynamicTransition struct {
	source               Event
	destinationFunc      func(context.Context) ([]Event, error)
	possibleDestinations []PossibleDestination
}

// NewDynamicTransition creates a new dynamic transition from the source event using the given destination function
// and possible destinations for documentation.
func NewDynamicTransition(source Event, destinationFunc func(context.Context) ([]Event, error), possibleDestinations []PossibleDestination) Transition {
	return &dynamicTransition{source: source, destinationFunc: destinationFunc, possibleDestinations: possibleDestinations}
}

// Source implements the Transition interface.
func (t *dynamicTransition) Source() Event {
	return t.source
}

// Destination implements the Transition interface.
func (t *dynamicTransition) Destination(ctx context.Context) ([]Event, error) {
	return t.destinationFunc(ctx)
}

// IsExclusive implements the Transition interface.
// Dynamic transitions are exclusive and prevent other transitions from being applied
func (t *dynamicTransition) IsExclusive() bool {
	return true
}

// PossibleDestinations implements the Transition interface.
func (t *dynamicTransition) PossibleDestinations() []PossibleDestination {
	return t.possibleDestinations
}
