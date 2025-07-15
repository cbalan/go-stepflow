package core

import "context"

type funcItem struct {
	scope        Scope
	activityFunc func(context.Context) error
}

func NewFuncItem(name string, activityFunc func(context.Context) error) StepFlowItem {
	return &funcItem{scope: NewScope(name), activityFunc: activityFunc}
}

func (fi *funcItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := WithParent(fi.scope, parent)

	destinationFunc := func(ctx context.Context) ([]Event, error) {
		if err := fi.activityFunc(ctx); err != nil {
			return nil, err
		}

		return []Event{CompletedEvent(scope)}, nil
	}

	return scope, []Transition{NewDynamicTransition(StartCommand(scope), destinationFunc)}, nil
}
