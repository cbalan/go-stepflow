package core

import "context"

type caseItem struct {
	scope         Scope
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewCaseItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &caseItem{
		scope:         NewScope(name),
		item:          item,
		conditionFunc: conditionFunc,
	}
}

func (ci *caseItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(ci.scope, parent)

	itemScope, itemTransitions, err := ci.item.Transitions(scope)
	if err != nil {
		return nil, nil, err
	}

	destinationFunc := func(ctx context.Context) ([]Event, error) {
		shouldStart, err := ci.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if shouldStart {
			return []Event{StartCommand(itemScope)}, nil
		}

		return []Event{CompletedEvent(scope)}, nil
	}

	transitions := []Transition{
		NewDynamicTransition(StartCommand(scope), destinationFunc, []PossibleDestination{
			NewReason(StartCommand(itemScope), "Case condition is met"),
			NewReason(CompletedEvent(scope), "Case condition is not met"),
		}),
		NewStaticTransition(CompletedEvent(itemScope), CompletedEvent(scope)),
	}

	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
