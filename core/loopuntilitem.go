package core

import "context"

type loopUntilItem struct {
	scope         Scope
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewLoopUntilItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &loopUntilItem{scope: NewScope(name), item: item, conditionFunc: conditionFunc}
}

func (lui *loopUntilItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(lui.scope, parent)

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
		NewDynamicTransition(CompletedEvent(itemScope), destinationFunc, []PossibleDestination{
			NewReason(StartCommand(itemScope), "LoopUntil condition is not met"),
			NewReason(CompletedEvent(scope), "LoopUntil condition is met"),
		}),
	}

	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
