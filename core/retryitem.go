package core

import "context"

// retriableTransition wraps another transition to provide retry behavior when errors occur.
// It delegates to an error handler function to determine whether to retry or propagate the error.
type retriableTransition struct {
	transition       Transition
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
	retryEvent       Event
}

// Source returns the source event of the wrapped transition.
func (rt *retriableTransition) Source() Event {
	return rt.transition.Source()
}

// Destination evaluates the wrapped transition and handles any errors by consulting the error handler.
func (rt *retriableTransition) Destination(ctx context.Context) ([]Event, error) {
	events, err := rt.transition.Destination(ctx)
	if err != nil {
		// If there's an error, consult the error handler.
		shouldRetry, errorHandlerErr := rt.errorHandlerFunc(ctx, err)
		if errorHandlerErr != nil {
			// If the error handler itself fails, propagate that error.
			return nil, errorHandlerErr
		}

		if shouldRetry {
			// If we should retry, transition to the retry event.
			return []Event{rt.retryEvent}, nil
		} else {
			// Otherwise, propagate the original error.
			return events, err
		}
	}

	// If there's no error, proceed normally
	return events, err
}

// IsExclusive delegates to the wrapped transition.
func (rt *retriableTransition) IsExclusive() bool {
	return rt.transition.IsExclusive()
}

// PossibleDestinations returns all possible destinations, including the retry event.
func (t *retriableTransition) PossibleDestinations() []PossibleDestination {
	var result []PossibleDestination
	result = append(result, t.transition.PossibleDestinations()...)
	result = append(result, NewReason(t.retryEvent, "retry"))
	return result
}

// retryItem wraps another workflow item to provide retry behavior when errors occur.
// It does not add any new events or transitions, but wraps the contained item's transitions
// to handle errors and potentially retry the item from the beginning.
type retryItem struct {
	item             StepFlowItem
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
}

// NewRetryItem creates a new workflow item that wraps another item with retry behavior.
// The error handler function receives the context and error, and should return true if
// the operation should be retried, or an error if the handler itself fails.
func NewRetryItem(item StepFlowItem, errorHandlerFunc func(ctx context.Context, err error) (bool, error)) StepFlowItem {
	return &retryItem{item: item, errorHandlerFunc: errorHandlerFunc}
}

// Transitions implements the StepFlowItem interface.
// It wraps each transition of the contained item with retry behavior.
func (ri *retryItem) Transitions(parent Scope) (Scope, []Transition, error) {
	// Get the item's scope and transitions.
	itemScope, itemTransitions, err := ri.item.Transitions(parent)
	if err != nil {
		return nil, nil, err
	}

	// Wrap each transition with retry behavior.
	var transitions []Transition
	for _, transition := range itemTransitions {
		transitions = append(transitions, &retriableTransition{
			transition:       transition,
			errorHandlerFunc: ri.errorHandlerFunc,
			retryEvent:       StartCommand(itemScope),
		})
	}

	return itemScope, transitions, nil
}
