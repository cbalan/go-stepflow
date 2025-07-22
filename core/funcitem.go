package core

import "context"

// funcItem represents a workflow item that executes a single function when started.
// It completes when the function returns successfully or returns an error if the function fails.
type funcItem struct {
	scope        Scope
	activityFunc func(context.Context) error
}

// NewFuncItem creates a new workflow item that executes the given function when started.
// The function receives a context and should return an error if it fails.
func NewFuncItem(name string, activityFunc func(context.Context) error) StepFlowItem {
	return &funcItem{scope: NewScope(name), activityFunc: activityFunc}
}

// Transitions implements the StepFlowItem interface.
// It defines a single transition from start to completion that executes the activity function.
func (fi *funcItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(fi.scope, parent)

	// When the item starts, execute the activity function and complete if successful.
	destinationFunc := func(ctx context.Context) ([]Event, error) {
		if err := fi.activityFunc(ctx); err != nil {
			return nil, err
		}

		return []Event{CompletedEvent(scope)}, nil
	}

	transitions := []Transition{
		NewDynamicTransition(StartCommand(scope), destinationFunc, []PossibleDestination{
			NewReason(CompletedEvent(scope), "completed"),
		}),
	}

	return scope, transitions, nil
}
