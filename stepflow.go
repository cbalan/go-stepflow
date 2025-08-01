package stepflow

import (
	"context"
	"github.com/cbalan/go-stepflow/core"
)

type StepFlow = core.StepFlow

// NewStepFlow creates a new executable workflow from steps specification.
func NewStepFlow(stepsSpec *StepsSpec) (StepFlow, error) {
	return core.NewStepFlow(core.NewStepsItem(stepsSpec.name, stepsSpec.items))
}

// StepsSpec holds the structure of the step flow.
type StepsSpec struct {
	name  string
	items []core.StepFlowItem
}

// Named creates and returns a new StepsSpec with the given name for structuring step-based workflows.
func Named(name string) *StepsSpec {
	return &StepsSpec{name: name}
}

// Steps creates a new empty steps specification.
// This is the starting point for defining a workflow using the fluent API.
func Steps() *StepsSpec {
	return Named("steps")
}

// Steps adds a nested group of steps to the workflow with the given name.
// The steps in stepSpec are executed sequentially as a single logical unit.
// This is useful for organizing complex workflows into logical components.
func (s *StepsSpec) Steps(name string, stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewStepsItem(name, stepsSpec.items))
	return s
}

// Do adds a step that executes a function when the workflow reaches this point.
// This is the primary way to add business logic to a workflow.
func (s *StepsSpec) Do(name string, activityFunc func(ctx context.Context) error) *StepsSpec {
	s.items = append(s.items, core.NewFuncItem(name, activityFunc))
	return s
}

// WaitFor adds a step that pauses the workflow until a specified condition is met.
// The condition function is evaluated repeatedly. The workflow only proceeds
// when the function returns true.
func (s *StepsSpec) WaitFor(name string, conditionFunc func(ctx context.Context) (bool, error)) *StepsSpec {
	s.items = append(s.items, core.NewWaitForItem(name+"WaitFor", conditionFunc))
	return s
}

// Retry adds retry logic to a group of steps.
// If any step in the group fails with an error, the error handler function is called
// to determine whether to retry the entire group of steps.
func (s *StepsSpec) Retry(name string, errHandlerFunc func(ctx context.Context, err error) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewRetryItem(core.NewStepsItem(name+"Retry", stepsSpec.items), errHandlerFunc))
	return s
}

// LoopUntil adds a step that repeats a group of steps until a condition is met.
// After each execution of the steps, the condition function is evaluated.
// If it returns true, the workflow proceeds to the next step. Otherwise, the steps are executed again.
func (s *StepsSpec) LoopUntil(name string, conditionFunc func(ctx context.Context) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewLoopUntilItem(name+"LoopUntil", core.NewStepsItem("steps", stepsSpec.items), conditionFunc))
	return s
}

// Case adds a step that conditionally executes a group of steps based on a condition.
// The child steps are executed only if the condition function returns true.
// If the condition function returns false, the case step is skipped and the workflow proceeds to the next step.
func (s *StepsSpec) Case(name string, conditionFunc func(ctx context.Context) (bool, error), stepsSpec *StepsSpec) *StepsSpec {
	s.items = append(s.items, core.NewCaseItem(name+"Case", core.NewStepsItem("steps", stepsSpec.items), conditionFunc))
	return s
}

// WithName sets the steps specification name. Information mainly used for the top level steps.
// Deprecated: Please use Named(name)
func (s *StepsSpec) WithName(name string) *StepsSpec {
	s.name = name
	return s
}

// Transitions returns the list of transitions as defined by the steps specification.
// This helper function enables consumers to inspect the underlying workflow state machine.
func Transitions(stepsSpec *StepsSpec) (core.Scope, []core.Transition, error) {
	return core.NewStepsItem(stepsSpec.name, stepsSpec.items).Transitions(nil)
}
