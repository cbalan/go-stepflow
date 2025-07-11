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
		transitionsMap[t.Source()] = append(transitionsMap[t.Source()], t)
	}

	return &stepFlow{item: item, scope: itemScope, transitionsMap: transitionsMap}, nil
}

const ApplyOneMaxIterations = 100

func (sf *stepFlow) Apply(ctx context.Context, oldState []string) ([]string, error) {
	newState := withDefaultValue(oldState, []string{StartCommand(sf.scope)})
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
			return newState, isExclusive, err
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

	return state[0] == CompletedEvent(sf.scope)
}

func eventString(event string, scope Scope) string {
	return event + ":" + scope.String()
}

func StartCommand(scope Scope) string {
	return eventString("start", scope)
}

func CompletedEvent(scope Scope) string {
	return eventString("completed", scope)
}

type Transition interface {
	Source() string
	Destination(context.Context) ([]string, error)
	IsExclusive() bool
}

type staticTransition struct {
	source      string
	destination string
}

func (t staticTransition) Source() string {
	return t.source
}

func (t staticTransition) Destination(_ context.Context) ([]string, error) {
	return []string{t.destination}, nil
}

func (t staticTransition) IsExclusive() bool {
	return false
}

type dynamicTransition struct {
	source          string
	destinationFunc func(context.Context) ([]string, error)
}

func (t dynamicTransition) Source() string {
	return t.source
}

func (t dynamicTransition) Destination(ctx context.Context) ([]string, error) {
	return t.destinationFunc(ctx)
}

func (t dynamicTransition) IsExclusive() bool {
	return true
}
