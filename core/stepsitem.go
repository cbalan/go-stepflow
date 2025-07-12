package core

import "fmt"

type ItemsProvider interface {
	Items() ([]StepFlowItem, error)
}

type stepsItem struct {
	name          string
	itemsProvider ItemsProvider
}

func NewStepsItem(name string, itemsProvider ItemsProvider) StepFlowItem {
	return &stepsItem{name: name, itemsProvider: itemsProvider}
}

func (si *stepsItem) Name() string {
	return si.name
}

func (si *stepsItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := NewItemScope(si, parent)

	items, err := si.itemsProvider.Items()
	if err != nil {
		return nil, nil, err
	}

	seenNames := make(map[string]bool)

	lastEvent := StartCommand(scope)

	var transitions []Transition
	for _, item := range items {
		itemScope, itemTransitions, err := item.Transitions(scope)
		if err != nil {
			return nil, nil, err
		}

		if seenNames[item.Name()] {
			return nil, nil, fmt.Errorf("name %s must be unique in the current context", item.Name())
		}
		seenNames[item.Name()] = true

		transitions = append(transitions, NewStaticTransition(lastEvent, StartCommand(itemScope)))

		transitions = append(transitions, itemTransitions...)

		lastEvent = CompletedEvent(itemScope)
	}

	transitions = append(transitions, NewStaticTransition(lastEvent, CompletedEvent(scope)))

	return scope, transitions, nil
}
