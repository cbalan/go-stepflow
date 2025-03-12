package core

import "context"

type retriableTransition struct {
	transition       Transition
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
	retryEvent       string
}

func (rt retriableTransition) Source() string {
	return rt.transition.Source()
}

func (rt retriableTransition) Destination(ctx context.Context) ([]string, error) {
	events, err := rt.transition.Destination(ctx)
	if err != nil {
		shouldRetry, errorHandlerErr := rt.errorHandlerFunc(ctx, err)
		if errorHandlerErr != nil {
			return nil, errorHandlerErr
		}

		if shouldRetry {
			return []string{rt.retryEvent}, nil
		} else {
			return events, err
		}
	}

	return events, err
}

func (rt retriableTransition) IsExclusive() bool {
	return rt.transition.IsExclusive()
}

type retryItem struct {
	item             StepFlowItem
	errorHandlerFunc func(ctx context.Context, err error) (bool, error)
}

func NewRetryItem(item StepFlowItem, errorHandlerFunc func(ctx context.Context, err error) (bool, error)) StepFlowItem {
	return &retryItem{item: item, errorHandlerFunc: errorHandlerFunc}
}

func (ri *retryItem) Name() string {
	return ri.item.Name()
}

func (ri *retryItem) WithName(name string) StepFlowItem {
	return &retryItem{
		item:             ri.item.WithName(name),
		errorHandlerFunc: ri.errorHandlerFunc,
	}
}

func (ri *retryItem) Transitions() ([]Transition, error) {
	itemTransitions, err := ri.item.Transitions()
	if err != nil {
		return nil, err
	}

	var transitions []Transition
	for _, transition := range itemTransitions {
		transitions = append(transitions, retriableTransition{
			transition:       transition,
			errorHandlerFunc: ri.errorHandlerFunc,
			retryEvent:       StartCommand(ri.item),
		})
	}

	return transitions, nil
}
