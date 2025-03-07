package stepflow

import "context"

type caseItem struct {
	name          string
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func newCaseItem(item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &caseItem{
		item:          item,
		conditionFunc: conditionFunc,
	}
}

func (ci *caseItem) Name() string {
	return ci.name
}

func (ci *caseItem) WithName(name string) StepFlowItem {
	return &caseItem{
		name:          name,
		item:          ci.item.WithName(namespacedName(name, "case")),
		conditionFunc: ci.conditionFunc,
	}
}

func (ci *caseItem) Transitions() ([]Transition, error) {
	transitions := []Transition{
		dynamicTransition{source: StartCommand(ci), destinationFunc: ci.apply},
		staticTransition{source: CompletedEvent(ci.item), destination: CompletedEvent(ci)},
	}

	itemTransitions, err := ci.item.Transitions()
	if err != nil {
		return nil, err
	}

	transitions = append(transitions, itemTransitions...)

	return transitions, nil
}

func (ci *caseItem) apply(ctx context.Context) ([]string, error) {
	shouldStart, err := ci.conditionFunc(ctx)
	if err != nil {
		return nil, err
	}

	if shouldStart {
		return []string{StartCommand(ci.item)}, nil
	}

	return []string{CompletedEvent(ci)}, nil
}
