package core

import "fmt"

type stepsItem struct {
	scope Scope
	items []StepFlowItem
}

func NewStepsItem(name string, items []StepFlowItem) StepFlowItem {
	return &stepsItem{scope: NewScope(name), items: items}
}

func (si *stepsItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(si.scope, parent)

	seenNames := make(map[string]bool)

	lastEvent := StartCommand(scope)

	var transitions []Transition

	for _, item := range si.items {
		itemScope, itemTransitions, err := item.Transitions(scope)
		if err != nil {
			return nil, nil, err
		}

		if seenNames[itemScope.Name()] {
			return nil, nil, fmt.Errorf("name %s must be unique in the current context", itemScope.Name())
		}
		seenNames[itemScope.Name()] = true

		transitions = append(transitions, NewStaticTransition(lastEvent, StartCommand(itemScope)))

		transitions = append(transitions, itemTransitions...)

		lastEvent = CompletedEvent(itemScope)
	}

	transitions = append(transitions, NewStaticTransition(lastEvent, CompletedEvent(scope)))

	return scope, transitions, nil
}
