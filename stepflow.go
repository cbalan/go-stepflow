package stepflow

import (
	"context"
	"github.com/cbalan/go-stepflow/core"
)

type StepFlow = core.StepFlow

type StepSpec interface {
	Steps(name string, stepSpec StepSpec) StepSpec
	Do(name string, activityFunc func(ctx context.Context) error) StepSpec
	WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec
	Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepSpec StepSpec) StepSpec
	LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec
	Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec
}

type stepSpecImpl []core.StepFlowItem

// Items implements core.ItemsProvider interface.
func (s stepSpecImpl) Items() ([]core.StepFlowItem, error) {
	return s, nil
}

// Steps implements StepSpec interface.
func (s stepSpecImpl) Steps(name string, stepSpec StepSpec) StepSpec {
	return append(s, core.NewStepsItem(name, stepSpec.(core.ItemsProvider)))
}

// Do implements StepSpec interface
func (s stepSpecImpl) Do(name string, activityFunc func(ctx context.Context) error) StepSpec {
	return append(s, core.NewFuncItem(name, activityFunc))
}

// WaitFor implements StepSpec interface
func (s stepSpecImpl) WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec {
	return append(s, core.NewWaitForItem(name+"WaitFor", conditionFunc))
}

// Retry implements StepSpec interface
func (s stepSpecImpl) Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewRetryItem(core.NewStepsItem(name+"Retry", stepSpec.(core.ItemsProvider)), errHandlerFunc))
}

// LoopUntil implements StepSpec interface
func (s stepSpecImpl) LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewLoopUntilItem(name+"LoopUntil", core.NewStepsItem("steps", stepSpec.(core.ItemsProvider)), conditionFunc))
}

// Case implements StepSpec interface
func (s stepSpecImpl) Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewCaseItem(name+"Case", core.NewStepsItem("steps", stepSpec.(core.ItemsProvider)), conditionFunc))
}

func Steps() StepSpec {
	return stepSpecImpl{}
}

func NewStepFlow(name string, stepSpec StepSpec) (StepFlow, error) {
	return core.NewStepFlow(core.NewStepsItem(name, stepSpec.(core.ItemsProvider)))
}
