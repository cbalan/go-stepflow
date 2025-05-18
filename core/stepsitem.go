package core

import "fmt"

type ItemsProvider interface {
	Items() ([]StepFlowItem, error)
}

type stepsItem struct {
	name          string
	itemsProvider ItemsProvider
}

func NewStepsItem(itemsProvider ItemsProvider) StepFlowItem {
	return &stepsItem{itemsProvider: itemsProvider}
}

func (si *stepsItem) Name() string {
	return si.name
}

func (si *stepsItem) WithName(name string) StepFlowItem {
	return &stepsItem{name: name, itemsProvider: si.itemsProvider}
}

func (si *stepsItem) Transitions() ([]Transition, error) {
	items, err := si.itemsProvider.Items()
	if err != nil {
		return nil, err
	}

	seenNames := make(map[string]bool)

	lastEvent := StartCommand(si)

	var transitions []Transition
	for _, item := range items {
		nsItem := item.WithName(NamespacedName(si.Name(), item.Name()))

		if seenNames[nsItem.Name()] {
			return nil, fmt.Errorf("name %s must be unique in the current context", nsItem.Name())
		}
		seenNames[nsItem.Name()] = true

		transitions = append(transitions, staticTransition{source: lastEvent, destination: StartCommand(nsItem)})

		itemTransitions, err := nsItem.Transitions()
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, itemTransitions...)

		lastEvent = CompletedEvent(nsItem)
	}

	transitions = append(transitions, staticTransition{source: lastEvent, destination: CompletedEvent(si)})

	return transitions, nil
}
