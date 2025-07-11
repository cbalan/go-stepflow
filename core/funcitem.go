package core

import "context"

type funcItem struct {
	name         string
	activityFunc func(context.Context) error
}

func NewFuncItem(name string, activityFunc func(context.Context) error) StepFlowItem {
	return &funcItem{name: name, activityFunc: activityFunc}
}

func (fi *funcItem) Name() string {
	return fi.name
}

func (fi *funcItem) Transitions(parent Scope) (Scope, []Transition, error) {
	scope := NewItemScope(fi, parent)

	destinationFunc := func(ctx context.Context) ([]string, error) {
		if err := fi.activityFunc(ctx); err != nil {
			return nil, err
		}

		return []string{CompletedEvent(scope)}, nil
	}

	return scope, []Transition{dynamicTransition{source: StartCommand(scope), destinationFunc: destinationFunc}}, nil
}
