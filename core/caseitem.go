package core

import "context"

type caseItem struct {
	name          string
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func NewCaseItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &caseItem{
		name:          name,
		item:          item,
		conditionFunc: conditionFunc,
	}
}

func (ci *caseItem) Name() string {
	return ci.name
}

func (ci *caseItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := NewItemScope(ci, parent)

	itemScope, itemTransitions, err := ci.item.Transitions(scope)
	if err != nil {
		return nil, nil, err
	}

	destinationFunc := func(ctx context.Context) ([]string, error) {
		shouldStart, err := ci.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if shouldStart {
			return []string{StartCommand(itemScope)}, nil
		}

		return []string{CompletedEvent(scope)}, nil
	}

	transitions := []Transition{
		dynamicTransition{source: StartCommand(scope), destinationFunc: destinationFunc},
		staticTransition{source: CompletedEvent(itemScope), destination: CompletedEvent(scope)},
	}

	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
