package core

import "context"

type loopUntilItem struct {
	name          string
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewLoopUntilItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &loopUntilItem{name: name, item: item, conditionFunc: conditionFunc}
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

	destinationFunc := func(ctx context.Context) ([]Event, error) {
		completed, err := lui.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			return []Event{CompletedEvent(scope)}, nil
		}

		return []Event{StartCommand(itemScope)}, nil
	}

	transitions := []Transition{
		NewStaticTransition(StartCommand(scope), StartCommand(itemScope)),
		NewDynamicTransition(CompletedEvent(itemScope), destinationFunc),
	}

	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
