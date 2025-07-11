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

	destinationFunc := func(ctx context.Context) ([]string, error) {
		completed, err := wfi.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			return []string{CompletedEvent(scope)}, nil
		}

		return []string{StartCommand(scope)}, nil
	}

	transitions := []Transition{dynamicTransition{source: StartCommand(scope), destinationFunc: destinationFunc}}

	return scope, transitions, nil
}
