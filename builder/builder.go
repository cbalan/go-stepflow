package builder

import (
	"context"
	"github.com/cbalan/go-stepflow/core"
)

type StepFlow = core.StepFlow
type StepFlowItem = core.StepFlowItem

type StepFlowItemizer interface {
	StepFlowItem() (StepFlowItem, error)
}

type StepSpec interface {
	StepFlowItem() (StepFlowItem, error)
	Steps(name string, stepSpec StepSpec) StepSpec
	Do(name string, activityFunc func(ctx context.Context) error) StepSpec
	WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec
	Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), item StepFlowItemizer) StepSpec
	LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), item StepFlowItemizer) StepSpec
	Case(name string, conditionFunc func(ctx context.Context) (bool, error), item StepFlowItemizer) StepSpec
}

type stepSpecImpl struct {
	items []StepFlowItemizer
}

type stepFlowItemizerImpl struct {
	stepFlowItemFunc func() (StepFlowItem, error)
}

func (s *stepFlowItemizerImpl) StepFlowItem() (StepFlowItem, error) {
	return s.stepFlowItemFunc()
}

// Items implements core.ItemsProvider interface.
func (s *stepSpecImpl) Items(namespace string) ([]StepFlowItem, error) {
	var items []StepFlowItem

	for _, item := range s.items {
		actualItem, err := item.StepFlowItem()
		if err != nil {
			return nil, err
		}

		items = append(items, actualItem.WithName(core.NamespacedName(namespace, actualItem.Name())))
	}

	return items, nil
}

// StepFlowItem implements StepFlowItemizer interface.
func (s *stepSpecImpl) StepFlowItem() (StepFlowItem, error) {
	return core.NewStepsItem(s), nil
}

// Steps implements StepSpec interface
func (s *stepSpecImpl) Steps(name string, stepSpec StepSpec) StepSpec {
	stepsItem := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		actualItem, err := stepSpec.StepFlowItem()
		if err != nil {
			return nil, err
		}

		return actualItem.WithName(name), nil
	}}

	s.items = append(s.items, stepsItem)
	return s
}

// Do implements StepSpec interface
func (s *stepSpecImpl) Do(name string, activityFunc func(ctx context.Context) error) StepSpec {
	item := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		return core.NewFuncItem(activityFunc).WithName(name), nil
	}}

	s.items = append(s.items, item)
	return s
}

// WaitFor implements StepSpec interface
func (s *stepSpecImpl) WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec {
	item := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		return core.NewWaitForItem(conditionFunc).WithName(name), nil
	}}

	s.items = append(s.items, item)
	return s
}

// Retry implements StepSpec interface
func (s *stepSpecImpl) Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), item StepFlowItemizer) StepSpec {
	retryItem := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		actualItem, err := item.StepFlowItem()
		if err != nil {
			return nil, err
		}

		return core.NewRetryItem(actualItem, errHandlerFunc).WithName(name), nil
	}}

	s.items = append(s.items, retryItem)
	return s
}

// LoopUntil implements StepSpec interface
func (s *stepSpecImpl) LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), item StepFlowItemizer) StepSpec {
	loopUntilItem := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		actualItem, err := item.StepFlowItem()
		if err != nil {
			return nil, err
		}

		return core.NewLoopUntilItem(actualItem, conditionFunc).WithName(name), nil
	}}

	s.items = append(s.items, loopUntilItem)
	return s
}

// Case implements StepSpec interface
func (s *stepSpecImpl) Case(name string, conditionFunc func(ctx context.Context) (bool, error), item StepFlowItemizer) StepSpec {
	caseItem := &stepFlowItemizerImpl{func() (StepFlowItem, error) {
		actualItem, err := item.StepFlowItem()
		if err != nil {
			return nil, err
		}

		return core.NewCaseItem(actualItem, conditionFunc).WithName(name), nil
	}}

	s.items = append(s.items, caseItem)
	return s
}

func Steps() StepSpec {
	return &stepSpecImpl{}
}

func NewStepFlow(name string, item StepFlowItemizer) (StepFlow, error) {
	actualItem, err := item.StepFlowItem()
	if err != nil {
		return nil, err
	}

	return core.NewStepFlow(actualItem.WithName(name))
}
