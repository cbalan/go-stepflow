package stepflow

import (
	"context"
	"github.com/cbalan/go-stepflow/core"
)

type StepsSpec struct {
	items []core.StepFlowItem
}

// Items returns the core.StepFlowItem items defined via StepsSpec.
func Items(stepsSpec *StepsSpec) []core.StepFlowItem {
	return stepsSpec.items
}

func (s *StepsSpec) Steps(name string, stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewStepsItem(name, Items(stepsSpec)))
	return s
}

func (s *StepsSpec) Do(name string, activityFunc func(ctx context.Context) error) *StepsSpec {
	s.items = append(s.items, core.NewFuncItem(name, activityFunc))
	return s
}

func (s *StepsSpec) WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) *StepsSpec {
	s.items = append(s.items, core.NewWaitForItem(name+"WaitFor", conditionFunc))
	return s
}

func (s *StepsSpec) Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewRetryItem(core.NewStepsItem(name+"Retry", Items(stepsSpec)), errHandlerFunc))
	return s
}

func (s *StepsSpec) LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewLoopUntilItem(name+"LoopUntil", core.NewStepsItem("steps", Items(stepsSpec)), conditionFunc))
	return s
}

func (s *StepsSpec) Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewCaseItem(name+"Case", core.NewStepsItem("steps", Items(stepsSpec)), conditionFunc))
	return s
}

func Steps() *StepsSpec {
	return &StepsSpec{}
}

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

func NewStepFlow(name string, stepsSpec *StepsSpec) (StepFlow, error) {
	return core.NewStepFlow(core.NewStepsItem(name, Items(stepsSpec)))
}
