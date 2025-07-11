package core

import "context"

type loopUntilItem struct {
	name          string
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewLoopUntilItem(item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &loopUntilItem{item: item, conditionFunc: conditionFunc}
}

func (lui *loopUntilItem) Name() string {
	return lui.name
}

func (lui *loopUntilItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := NewItemScope(lui, parent)

	itemScope, itemTransitions, err := lui.item.Transitions(scope)
	if err != nil {
		return nil, nil, err
	}

	destinationFunc := func(ctx context.Context) ([]string, error) {
		completed, err := lui.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			return []string{CompletedEvent(scope)}, nil
		}

		return []string{StartCommand(itemScope)}, nil
	}

	transitions := []Transition{
		staticTransition{source: StartCommand(scope), destination: StartCommand(itemScope)},
		dynamicTransition{source: CompletedEvent(itemScope), destinationFunc: destinationFunc},
	}

	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
