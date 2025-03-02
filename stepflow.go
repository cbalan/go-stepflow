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
	rootItem       StepFlowItem
	transitionsMap map[string][]Transition
}

func NewStepFlow(name string, nameItemPairs ...any) (StepFlow, error) {
	rootItem := Steps(nameItemPairs...).WithName(name)

	transitions, err := rootItem.Transitions()
	if err != nil {
		return nil, err
	}

	transitionsMap := make(map[string][]Transition)
	for _, t := range transitions {
		transitionsMap[t.Source()] = append(transitionsMap[t.Source()], t)
	}

	return &stepFlow{rootItem: rootItem, transitionsMap: transitionsMap}, nil
}

func (sf *stepFlow) Apply(ctx context.Context, state []string) ([]string, error) {
	if sf.IsCompleted(state) {
		return state, nil
	}

	stateWithDefault := state
	if stateWithDefault == nil {
		stateWithDefault = []string{StartCommand(sf.rootItem)}
	}

	for _, lastEvent := range stateWithDefault {
		for _, t := range sf.transitionsMap[lastEvent] {
			return t.Destination(ctx)
		}
	}

	return nil, fmt.Errorf("unhnadled state %s", state)
}

func (sf *stepFlow) IsCompleted(state []string) bool {
	if len(state) != 1 {
		return false
	}

	return state[0] == CompletedEvent(sf.rootItem)
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
