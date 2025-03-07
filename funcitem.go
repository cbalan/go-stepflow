package stepflow

import "context"

type funcItem struct {
	name         string
	activityFunc func(context.Context) error
}

func newFuncItem(activityFunc func(context.Context) error) StepFlowItem {
	return &funcItem{activityFunc: activityFunc}
}

func (fi *funcItem) Name() string {
	return fi.name
}

func (fi *funcItem) WithName(name string) StepFlowItem {
	return &funcItem{name: name, activityFunc: fi.activityFunc}
}

func (fi *funcItem) Transitions() ([]Transition, error) {
	return []Transition{dynamicTransition{source: StartCommand(fi), destinationFunc: fi.apply}}, nil
}

func (fi *funcItem) apply(ctx context.Context) ([]string, error) {
	if err := fi.activityFunc(ctx); err != nil {
		return nil, err
	}

	return []string{CompletedEvent(fi)}, nil
}
