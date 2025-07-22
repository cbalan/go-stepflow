package core

import "context"

// waitForItem represents a workflow item that waits until a condition is met before completing.
// The condition is evaluated when the item starts, and if not met, the item continues waiting.
type waitForItem struct {
	scope         Scope
	conditionFunc func(ctx context.Context) (bool, error)
}

// NewWaitForItem creates a new workflow item that waits until the given condition function returns true.
// The condition function receives a context and should return true when the wait is complete,
// or an error if the evaluation fails.
func NewWaitForItem(name string, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &waitForItem{scope: NewScope(name), conditionFunc: conditionFunc}
}

// Transitions implements the StepFlowItem interface.
// It defines a self-transition that repeatedly evaluates the condition until it returns true.
func (wfi *waitForItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(wfi.scope, parent)

	// When the item starts, evaluate the condition.
	destinationFunc := func(ctx context.Context) ([]Event, error) {
		completed, err := wfi.conditionFunc(ctx)
		if err != nil {
			return nil, err
		}

		if completed {
			// Condition is met, complete the item.
			return []Event{CompletedEvent(scope)}, nil
		}

		// Condition is not met, continue waiting.
		return []Event{StartCommand(scope)}, nil
	}

	transitions := []Transition{
		NewDynamicTransition(StartCommand(scope), destinationFunc, []PossibleDestination{
			NewReason(StartCommand(scope), "WaitFor condition is not met"),
			NewReason(CompletedEvent(scope), "WaitFor condition is met"),
		}),
	}

	return scope, transitions, nil
}
