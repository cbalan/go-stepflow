package core

import (
	"context"
	"fmt"
	"strings"
)

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

type Scope interface {
	Parent() Scope
	Name() string
}

type scopeImpl struct {
	name   string
	parent Scope
}

func NewItemScope(item StepFlowItem, parent Scope) Scope {
	return &scopeImpl{parent: parent, name: item.Name()}
}

func (s *scopeImpl) Name() string {
	return s.name
}

func (s *scopeImpl) Parent() Scope {
	return s.parent
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
	var sb strings.Builder

	sb.WriteString(event)
	sb.WriteString(":")

	names := scopeNames(scope)
	if len(names) > 0 {
		sb.WriteString(names[len(names)-1])
		if len(names) > 1 {
			for i := len(names) - 2; i >= 0; i-- {
				sb.WriteString("/")
				sb.WriteString(names[i])
			}
		}
	}

	return sb.String()
}

func scopeNames(scope Scope) []string {
	var names []string
	for scope != nil {
		names = append(names, scope.Name())
		scope = scope.Parent()
	}
	return names
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

func NamespacedName(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
