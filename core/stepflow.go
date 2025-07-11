package core

import (
	"context"
	"fmt"
)

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

type Scope interface {
	String() string
}

type scopeImpl string

func NewItemScope(item StepFlowItem, parent Scope) Scope {
	if parent == nil {
		return scopeImpl(item.Name())
	}

	return scopeImpl(parent.String() + "/" + item.Name())
}

func (s scopeImpl) String() string {
	return string(s)
}

type StepFlowItem interface {
	Name() string
	Transitions(parent Scope) (Scope, []Transition, error)
}

type stepFlow struct {
	item           StepFlowItem
	scope          Scope
	transitionsMap map[string][]Transition
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

	return &stepFlow{item: item, scope: itemScope, transitionsMap: transitionsMap}, nil
}

const ApplyOneMaxIterations = 100

func (sf *stepFlow) Apply(ctx context.Context, oldState []string) ([]string, error) {
	newState := withDefaultValue(oldState, []string{eventString(StartCommand(sf.scope))})
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

func (sf *stepFlow) applyOne(ctx context.Context, oldState []string) ([]string, bool, error) {
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

func (sf *stepFlow) IsCompleted(state []string) bool {
	if len(state) != 1 {
		return false
	}

	return state[0] == eventString(CompletedEvent(sf.scope))
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
	return event.Name() + ":" + event.Scope().String()
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

type Transition interface {
	Source() Event
	Destination(context.Context) ([]Event, error)
	IsExclusive() bool
}

type staticTransition struct {
	source      Event
	destination []Event
}

func NewStaticTransition(source Event, destination ...Event) Transition {
	return staticTransition{source: source, destination: destination}
}

func (t staticTransition) Source() Event {
	return t.source
}

func (t staticTransition) Destination(_ context.Context) ([]Event, error) {
	return t.destination, nil
}

func (t staticTransition) IsExclusive() bool {
	return false
}

type dynamicTransition struct {
	source          Event
	destinationFunc func(context.Context) ([]Event, error)
}

func NewDynamicTransition(source Event, destinationFunc func(context.Context) ([]Event, error)) Transition {
	return dynamicTransition{source: source, destinationFunc: destinationFunc}
}

func (t dynamicTransition) Source() Event {
	return t.source
}

func (t dynamicTransition) Destination(ctx context.Context) ([]Event, error) {
	return t.destinationFunc(ctx)
}

func (t dynamicTransition) IsExclusive() bool {
	return true
}
