package stepflow

import "context"

type loopUntilItem struct {
	name          string
	item          StepFlowItem
	conditionFunc func(ctx context.Context) (bool, error)
}

func newLoopUntilItem(item StepFlowItem, conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &loopUntilItem{item: item, conditionFunc: conditionFunc}
}

func (lui *loopUntilItem) Name() string {
	return lui.name
}

func (lui *loopUntilItem) WithName(name string) StepFlowItem {
	return &loopUntilItem{
		name:          name,
		item:          lui.item.WithName(namespacedName(name, "loop")),
		conditionFunc: lui.conditionFunc,
	}
}

func (lui *loopUntilItem) Transitions() ([]Transition, error) {
	transitions := []Transition{
		staticTransition{source: StartCommand(lui), destination: StartCommand(lui.item)},
		dynamicTransition{source: CompletedEvent(lui.item), destinationFunc: lui.apply},
	}

	loopStepsTransitions, err := lui.item.Transitions()
	if err != nil {
		return nil, err
	}

	transitions = append(transitions, loopStepsTransitions...)

	return transitions, nil
}

func (lui *loopUntilItem) apply(ctx context.Context) ([]string, error) {
	completed, err := lui.conditionFunc(ctx)
	if err != nil {
		return nil, err
	}

	if completed {
		return []string{CompletedEvent(lui)}, nil
	}

	return []string{StartCommand(lui.item)}, nil
}
