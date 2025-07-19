package core

import "context"

type retriableTransition struct {
	transition       Transition
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
	retryEvent       Event
}

func (rt *retriableTransition) Source() Event {
	return rt.transition.Source()
}

func (rt *retriableTransition) Destination(ctx context.Context) ([]Event, error) {
	events, err := rt.transition.Destination(ctx)
	if err != nil {
		shouldRetry, errorHandlerErr := rt.errorHandlerFunc(ctx, err)
		if errorHandlerErr != nil {
			return nil, errorHandlerErr
		}

		if shouldRetry {
			return []Event{rt.retryEvent}, nil
		} else {
			return events, err
		}
	}

	return events, err
}

func (rt *retriableTransition) IsExclusive() bool {
	return rt.transition.IsExclusive()
}

type retryItem struct {
	item             StepFlowItem
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
}

func NewRetryItem(item StepFlowItem, errorHandlerFunc func(ctx context.Context, err error) (bool, error)) StepFlowItem {
	return &retryItem{item: item, errorHandlerFunc: errorHandlerFunc}
}

func (ri *retryItem) Transitions(parent Scope) (Scope, []Transition, error) {
	itemScope, itemTransitions, err := ri.item.Transitions(parent)
	if err != nil {
		return nil, nil, err
	}

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
