package stepflow

import "context"

func NewStepFlow(name string, nameItemPairs ...any) (StepFlow, error) {
	return newStepFlow(newStepsItem(nameItemPairs).WithName(name))
}

func Steps(nameItemPairs ...any) StepFlowItem {
	return newStepsItem(nameItemPairs)
}

func Case(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newCaseItem(newStepsItem(nameItemPairs), conditionFunc)
}

func LoopUntil(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newLoopUntilItem(newStepsItem(nameItemPairs), conditionFunc)
}

func Retry(errorHandlerFunc func(ctx context.Context, err error) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newRetryItem(newStepsItem(nameItemPairs), errorHandlerFunc)
}

func WaitFor(conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return newWaitForItem(conditionFunc)
}
