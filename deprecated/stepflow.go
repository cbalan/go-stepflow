package deprecated

import (
	"context"
	"fmt"

	"github.com/cbalan/go-stepflow/core"
)

type StepFlow = core.StepFlow
type StepFlowItem = core.StepFlowItem

func NewStepFlow(name string, nameItemPairs ...any) (StepFlow, error) {
	return core.NewStepFlow(Steps(nameItemPairs...).WithName(name))
}

func Steps(nameItemPairs ...any) StepFlowItem {
	return core.NewStepsItem(newNameItemPairsProvider(nameItemPairs))
}

func Case(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return core.NewCaseItem(Steps(nameItemPairs...), conditionFunc)
}

func LoopUntil(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return core.NewLoopUntilItem((Steps(nameItemPairs...)), conditionFunc)
}

func Retry(errorHandlerFunc func(ctx context.Context, err error) (bool, error), nameItemPairs ...any) StepFlowItem {
	return core.NewRetryItem(Steps(nameItemPairs...), errorHandlerFunc)
}

func WaitFor(conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return core.NewWaitForItem(conditionFunc)
}

type nameItemPairsProvider struct {
	nameItemPairs []any
}

func newNameItemPairsProvider(nameItemPairs []any) *nameItemPairsProvider {
	return &nameItemPairsProvider{nameItemPairs: nameItemPairs}
}

func (ni *nameItemPairsProvider) Items(namespace string) ([]StepFlowItem, error) {
	if len(ni.nameItemPairs)%2 != 0 {
		return nil, fmt.Errorf("un-even nameItemsPair")
	}

	seenNames := make(map[string]bool)

	var items []StepFlowItem
	for i := 0; i < len(ni.nameItemPairs); i += 2 {
		maybeName := ni.nameItemPairs[i]
		maybeItem := ni.nameItemPairs[i+1]

		name, ok := maybeName.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T used as string", maybeName)
		}

		if seenNames[name] {
			return nil, fmt.Errorf("name %s must be unique in the current context", name)
		}
		seenNames[name] = true

		item, err := newNamedItem(core.NamespacedName(namespace, name), maybeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create new named step flow item due to %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func newNamedItem(name string, maybeItem any) (StepFlowItem, error) {
	switch maybeItemV := maybeItem.(type) {
	case StepFlowItem:
		return maybeItemV.WithName(name), nil

	case func(context.Context) error:
		return core.NewFuncItem(maybeItemV).WithName(name), nil

	default:
		return nil, fmt.Errorf("type %T is not supported", maybeItemV)
	}
}
