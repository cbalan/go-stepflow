package stepflow

import (
	"context"
	"fmt"
)

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

type StepFlowItem interface {
	Name() string

	// WithName returns a new StepFlowItem instance with the provided name.
	WithName(string) StepFlowItem

	Transitions() ([]Transition, error)
}

type stepFlow struct {
	item           StepFlowItem
	transitionsMap map[string][]Transition
}

func newStepFlow(item StepFlowItem) (StepFlow, error) {
	transitions, err := item.Transitions()
	if err != nil {
		return nil, err
	}

	transitionsMap := make(map[string][]Transition)
	for _, t := range transitions {
		transitionsMap[t.Source()] = append(transitionsMap[t.Source()], t)
	}

	return &stepFlow{item: item, transitionsMap: transitionsMap}, nil
}

const ApplyOneMaxIterations = 100

func (sf *stepFlow) Apply(ctx context.Context, oldState []string) ([]string, error) {
	newState := withDefaultValue(oldState, []string{StartCommand(sf.item)})
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

	return state[0] == CompletedEvent(sf.item)
}

func eventString(item StepFlowItem, event string) string {
	return fmt.Sprintf("%s:%s", event, item.Name())
}

func StartCommand(item StepFlowItem) string {
	return eventString(item, "start")
}

func CompletedEvent(item StepFlowItem) string {
	return eventString(item, "completed")
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
