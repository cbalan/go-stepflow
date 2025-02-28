package stepflow

import (
	"context"
	"fmt"
)

type StepFlow struct {
	rootItem       StepFlowItem
	transitionsMap map[string][]Transition
}

type StepFlowItem interface {
	Name() string

	// WithName returns a new StepFlowItem instance with the provided name.
	WithName(string) StepFlowItem

	Transitions() ([]Transition, error)
}

func NewNamedStepFlow(name string, nameItemPairs ...any) (*StepFlow, error) {
	rootItem := Steps(nameItemPairs...).WithName(name)

	transitions, err := rootItem.Transitions()
	if err != nil {
		return nil, err
	}

	transitionsMap := make(map[string][]Transition)
	for _, t := range transitions {
		transitionsMap[t.Source()] = append(transitionsMap[t.Source()], t)
	}

	return &StepFlow{rootItem: rootItem, transitionsMap: transitionsMap}, nil
}

func (sf *StepFlow) Apply(ctx context.Context, state []string) ([]string, error) {
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

func (sf *StepFlow) IsCompleted(state []string) bool {
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
