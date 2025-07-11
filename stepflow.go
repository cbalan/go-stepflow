package stepflow

import (
	"context"
	"github.com/cbalan/go-stepflow/core"
)

type StepFlow = core.StepFlow
type StepFlowItem = core.StepFlowItem

type StepSpec interface {
	core.ItemsProvider

	Steps(name string, stepSpec StepSpec) StepSpec
	Do(name string, activityFunc func(ctx context.Context) error) StepSpec
	WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec
	Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepSpec StepSpec) StepSpec
	LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec
	Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec
}

type stepSpecImpl []StepFlowItem

// Items implements core.ItemsProvider interface.
func (s stepSpecImpl) Items() ([]StepFlowItem, error) {
	return s, nil
}

// Steps implements StepSpec interface.
func (s stepSpecImpl) Steps(name string, stepSpec StepSpec) StepSpec {
	return append(s, core.NewStepsItem(name, stepSpec))
}

// Do implements StepSpec interface
func (s stepSpecImpl) Do(name string, activityFunc func(ctx context.Context) error) StepSpec {
	return append(s, core.NewFuncItem(name, activityFunc))
}

// WaitFor implements StepSpec interface
func (s stepSpecImpl) WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) StepSpec {
	return append(s, core.NewWaitForItem("waitFor"+name, conditionFunc))
}

// Retry implements StepSpec interface
func (s stepSpecImpl) Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewRetryItem(core.NewStepsItem("retry"+name, stepSpec), errHandlerFunc))
}

// LoopUntil implements StepSpec interface
func (s stepSpecImpl) LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewLoopUntilItem("loopUntil"+name, core.NewStepsItem("steps", stepSpec), conditionFunc))
}

// Case implements StepSpec interface
func (s stepSpecImpl) Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepSpec StepSpec) StepSpec {
	return append(s, core.NewCaseItem("case"+name, core.NewStepsItem("steps", stepSpec), conditionFunc))
}

func Steps() StepSpec {
	return stepSpecImpl{}
}

func NewStepFlow(name string, stepSpec StepSpec) (StepFlow, error) {
	return core.NewStepFlow(core.NewStepsItem(name, stepSpec))
}
