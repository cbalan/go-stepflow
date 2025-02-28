package stepflow

import (
	"context"
	"fmt"
)

type stepsItem struct {
	name          string
	nameItemPairs []any
}

func Steps(nameItemPairs ...any) StepFlowItem {
	return &stepsItem{nameItemPairs: nameItemPairs}
}

func (si *stepsItem) Name() string {
	return si.name
}

func (si *stepsItem) WithName(name string) StepFlowItem {
	return &stepsItem{name: name, nameItemPairs: si.nameItemPairs}
}

func (si *stepsItem) Transitions() ([]Transition, error) {
	items, err := itemsFromNameItemPairs(si.name, si.nameItemPairs)
	if err != nil {
		return nil, err
	}

	lastEvent := StartCommand(si)

	var transitions []Transition
	for _, item := range items {
		transitions = append(transitions, staticTransition{source: lastEvent, destination: StartCommand(item)})

		itemTransitions, err := item.Transitions()
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, itemTransitions...)

		lastEvent = CompletedEvent(item)
	}

	transitions = append(transitions, staticTransition{source: lastEvent, destination: CompletedEvent(si)})

	return transitions, nil
}

func itemsFromNameItemPairs(namespace string, nameItemPairs []any) ([]StepFlowItem, error) {
	if len(nameItemPairs)%2 != 0 {
		return nil, fmt.Errorf("un-even nameItemsPair")
	}

	var items []StepFlowItem
	for i := 0; i < len(nameItemPairs); i += 2 {
		maybeName := nameItemPairs[i]
		maybeItem := nameItemPairs[i+1]

		name, ok := maybeName.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T used as string", maybeName)
		}

		item, err := newNamedItem(namespacedName(namespace, name), maybeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create new named step flow item due to %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func namespacedName(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func newNamedItem(name string, maybeItem any) (StepFlowItem, error) {
	switch maybeItemV := maybeItem.(type) {
	case StepFlowItem:
		return maybeItemV.WithName(name), nil

	case func(context.Context) error:
		return &funcItem{name: name, activityFunc: maybeItemV}, nil

	default:
		return nil, fmt.Errorf("type %T is not supported", maybeItemV)
	}
}
