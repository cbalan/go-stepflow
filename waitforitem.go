package stepflow

import "context"

type waitForItem struct {
	name          string
	conditionFunc func(ctx context.Context) (bool, error)
}

func WaitFor(conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return &waitForItem{conditionFunc: conditionFunc}
}

func (wfi *waitForItem) Name() string {
	return wfi.name
}

func (wfi *waitForItem) WithName(name string) StepFlowItem {
	return &waitForItem{
		name:          name,
		conditionFunc: wfi.conditionFunc,
	}
}

func (wfi *waitForItem) Transitions() ([]Transition, error) {
	transitions := []Transition{
		dynamicTransition{source: StartCommand(wfi), destinationFunc: wfi.apply},
	}

	return transitions, nil
}

func (wfi *waitForItem) apply(ctx context.Context) ([]string, error) {
	completed, err := wfi.conditionFunc(ctx)
	if err != nil {
		return nil, err
	}

	if completed {
		return []string{CompletedEvent(wfi)}, nil
	}

	return []string{StartCommand(wfi)}, nil
}
