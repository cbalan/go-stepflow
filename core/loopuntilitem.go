package core

import "context"

// loopUntilItem represents a workflow item that repeatedly executes another item until a condition is met.
// The condition is evaluated after each execution of the contained item.
type loopUntilItem struct {
	scope         Scope
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

// NewLoopUntilItem creates a new workflow item that repeatedly executes the given item
// until the condition function returns true.
// The condition function receives a context and should return true when the loop should stop,
// or an error if the evaluation fails.
func NewLoopUntilItem(name string, item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &loopUntilItem{scope: NewScope(name), item: item, conditionFunc: conditionFunc}
}

// Transitions implements the StepFlowItem interface.
// It connects the loop start to the item start, and the item completion back to either
// the item start (to loop again) or the loop completion (when the condition is met).
func (lui *loopUntilItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(lui.scope, parent)

	// Get the item's scope and transitions.
	itemScope, itemTransitions, err := lui.item.Transitions(scope)
	if err != nil {
		return nil, nil, err
	}

	// When the item completes, evaluate the condition.
	destinationFunc := func(ctx context.Context) ([]Event, error) {
		completed, err := lui.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			// Condition is met, complete the loop.
			return []Event{CompletedEvent(scope)}, nil
		}

		// Condition is not met, execute the item again.
		return []Event{StartCommand(itemScope)}, nil
	}

	transitions := []Transition{
		// When the loop starts, start the item
		NewStaticTransition(StartCommand(scope), StartCommand(itemScope)),
		// When the item completes, evaluate the condition
		NewDynamicTransition(CompletedEvent(itemScope), destinationFunc, []PossibleDestination{
			NewReason(StartCommand(itemScope), "LoopUntil condition is not met"),
			NewReason(CompletedEvent(scope), "LoopUntil condition is met"),
		}),
	}

	// Add item transitions.
	transitions = append(transitions, itemTransitions...)

	return scope, transitions, nil
}
