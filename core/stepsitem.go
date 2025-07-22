package core

import "fmt"

// stepsItem represents a sequence of workflow items that are executed one after another.
// Each item starts when the previous one completes, and the sequence completes when the last item completes.
type stepsItem struct {
	scope Scope
	items []StepFlowItem
}

// NewStepsItem creates a new workflow item that executes the given items in sequence.
func NewStepsItem(name string, items []StepFlowItem) StepFlowItem {
	return &stepsItem{scope: NewScope(name), items: items}
}

// Transitions implements the StepFlowItem interface.
// It connects the items in sequence, making each one start when the previous one completes.
func (si *stepsItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(si.scope, parent)

	// Track seen names to ensure uniqueness within this scope.
	seenNames := make(map[string]bool)

	// Initially, we transition from our start event.
	lastEvent := StartCommand(scope)

	var transitions []Transition

	for _, item := range si.items {
		// Get the item's scope and transitions.
		itemScope, itemTransitions, err := item.Transitions(scope)
		if err != nil {
			return nil, nil, err
		}

		// Ensure uniqueness of item names.
		if seenNames[itemScope.Name()] {
			return nil, nil, fmt.Errorf("name %s must be unique in the current context", itemScope.Name())
		}
		seenNames[itemScope.Name()] = true

		// Connect the last event to the start of this item.
		transitions = append(transitions, NewStaticTransition(lastEvent, StartCommand(itemScope)))

		// Add all items transitions.
		transitions = append(transitions, itemTransitions...)

		// The next item will start when this one completes.
		lastEvent = CompletedEvent(itemScope)
	}

	// Connect the last item's completion to our completion.
	transitions = append(transitions, NewStaticTransition(lastEvent, CompletedEvent(scope)))

	return scope, transitions, nil
}
