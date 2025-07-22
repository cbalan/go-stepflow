package core

import "context"

// caseItem represents a conditional workflow item that only executes another item if a condition is met.
// If the condition is not met, the case completes immediately without executing the item.
type caseItem struct {
	scope         Scope
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

// NewCaseItem creates a new workflow item that conditionally executes the given item
// based on the result of the condition function.
// The condition function receives a context and should return true if the item should be executed,
// or an error if the evaluation fails.
func NewCaseItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &caseItem{
		scope:         NewScope(name),
		item:          item,
		conditionFunc: conditionFunc,
	}
}

// Transitions implements the StepFlowItem interface.
// It evaluates the condition when the case starts and either executes the item
// or completes immediately based on the result.
func (ci *caseItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(ci.scope, parent)

	// Get the item's scope and transitions
	itemScope, itemTransitions, err := ci.item.Transitions(scope)
	if err != nil {
		return nil, nil, err
	}

	// When the case starts, evaluate the condition
	destinationFunc := func(ctx context.Context) ([]Event, error) {
		shouldStart, err := ci.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if shouldStart {
			// Condition is met, execute the item
			return []Event{StartCommand(itemScope)}, nil
		}

		// Condition is not met, complete the case without executing the item
		return []Event{CompletedEvent(scope)}, nil
	}

	transitions := []Transition{
		// When the case starts, evaluate the condition.
		NewDynamicTransition(StartCommand(scope), destinationFunc, []PossibleDestination{
			NewReason(StartCommand(itemScope), "Case condition is met"),
			NewReason(CompletedEvent(scope), "Case condition is not met"),
		}),

		// When the item completes, complete the case.
		NewStaticTransition(CompletedEvent(itemScope), CompletedEvent(scope)),
	}

	// Add all transitions for the item
	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
