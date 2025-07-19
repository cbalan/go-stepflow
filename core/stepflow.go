package core

import (
	"context"
	"fmt"
)

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

type stepFlowImpl struct {
	item           StepFlowItem
	transitionsMap map[string][]Transition
	startState     []string
	completedState []string
}

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

const ApplyOneMaxIterations = 100

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

	return nil, true, fmt.Errorf("unhnadled state %s", oldState)
}

func withDefaultValue(value []string, defaultValue []string) []string {
	if value == nil {
		return defaultValue
	}

	return value
}

func (sf *stepFlowImpl) IsCompleted(state []string) bool {
	if len(state) != 1 {
		return false
	}

	return state[0] == sf.completedState[0]
}

type Scope interface {
	Name() string
}

type scopeImpl string

func NewScope(name string) Scope {
	return scopeImpl(name)
}

func WithParent(scope Scope, parent Scope) Scope {
	if parent == nil {
		return scope
	}

	return scopeImpl(parent.Name() + "/" + scope.Name())
}

func (s scopeImpl) Name() string {
	return string(s)
}

type Event interface {
	Name() string
	Scope() Scope
}

type eventImpl struct {
	name  string
	scope Scope
}

func NewEvent(name string, scope Scope) Event {
	return &eventImpl{name: name, scope: scope}
}

func (e eventImpl) Name() string {
	return e.name
}

func (e eventImpl) Scope() Scope {
	return e.scope
}

func eventString(event Event) string {
	return event.Name() + ":" + event.Scope().Name()
}

func eventsString(events []Event) []string {
	var result []string
	for _, event := range events {
		result = append(result, eventString(event))
	}

	return result
}

func StartCommand(scope Scope) Event {
	return NewEvent("start", scope)
}

func CompletedEvent(scope Scope) Event {
	return NewEvent("completed", scope)
}

type StepFlowItem interface {
	Transitions(parent Scope) (Scope, []Transition, error)
}

type PossibleDestination interface {
	Event() Event
	Reason() string
}

type eventAndReason struct {
	event  Event
	reason string
}

func NewReason(event Event, reason string) PossibleDestination {
	return &eventAndReason{event: event, reason: reason}
}

func (n eventAndReason) Event() Event {
	return n.event
}

func (n eventAndReason) Reason() string {
	return n.reason
}

type Transition interface {
	Source() Event
	Destination(context.Context) ([]Event, error)
	IsExclusive() bool
	PossibleDestinations() []PossibleDestination
}

type staticTransition struct {
	source      Event
	destination []Event
}

func NewStaticTransition(source Event, destination ...Event) Transition {
	return &staticTransition{source: source, destination: destination}
}

func (t *staticTransition) Source() Event {
	return t.source
}

func (t *staticTransition) Destination(_ context.Context) ([]Event, error) {
	return t.destination, nil
}

func (t *staticTransition) IsExclusive() bool {
	return false
}

func (t *staticTransition) PossibleDestinations() []PossibleDestination {
	var result []PossibleDestination
	for _, destination := range t.destination {
		result = append(result, NewReason(destination, "static"))
	}
	return result
}

type dynamicTransition struct {
	source               Event
	destinationFunc      func(context.Context) ([]Event, error)
	possibleDestinations []PossibleDestination
}

func NewDynamicTransition(source Event, destinationFunc func(context.Context) ([]Event, error), possibleDestinations []PossibleDestination) Transition {
	return &dynamicTransition{source: source, destinationFunc: destinationFunc, possibleDestinations: possibleDestinations}
}

func (t *dynamicTransition) Source() Event {
	return t.source
}

func (t *dynamicTransition) Destination(ctx context.Context) ([]Event, error) {
	return t.destinationFunc(ctx)
}

func (t *dynamicTransition) IsExclusive() bool {
	return true
}

func (t *dynamicTransition) PossibleDestinations() []PossibleDestination {
	return t.possibleDestinations
}
