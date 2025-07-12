package core

import "context"

type waitForItem struct {
	name          string
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewWaitForItem(name string, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &waitForItem{name: name, conditionFunc: conditionFunc}
}

func (wfi *waitForItem) Name() string {
	return wfi.name
}

func (wfi *waitForItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := NewItemScope(wfi, parent)

	destinationFunc := func(ctx context.Context) ([]Event, error) {
		completed, err := wfi.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			return []Event{CompletedEvent(scope)}, nil
		}

		return []Event{StartCommand(scope)}, nil
	}

	transitions := []Transition{NewDynamicTransition(StartCommand(scope), destinationFunc)}

	return scope, transitions, nil
}
