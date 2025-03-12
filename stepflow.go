package stepflow

import (
	"context"
	"fmt"
)

type StepFlow interface {
	Apply(ctx context.Context, state []string) ([]string, error)
	IsCompleted(state []string) bool
}

type StepFlowItem interface {
	Name() string

	// WithName returns a new StepFlowItem instance with the provided name.
	WithName(string) StepFlowItem

	Transitions() ([]Transition, error)
}

func NewStepFlow(name string, nameItemPairs ...any) (StepFlow, error) {
	return newStepFlow(Steps(nameItemPairs...).WithName(name))
}

func Steps(nameItemPairs ...any) StepFlowItem {
	return newStepsItem(newNameItemPairsProvider(nameItemPairs))
}

func Case(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newCaseItem(Steps(nameItemPairs...), conditionFunc)
}

func LoopUntil(conditionFunc func(ctx context.Context) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newLoopUntilItem((Steps(nameItemPairs...)), conditionFunc)
}

func Retry(errorHandlerFunc func(ctx context.Context, err error) (bool, error), nameItemPairs ...any) StepFlowItem {
	return newRetryItem(Steps(nameItemPairs...), errorHandlerFunc)
}

func WaitFor(conditionFunc func(ctx context.Context) (bool, error)) StepFlowItem {
	return newWaitForItem(conditionFunc)
}

type stepFlow struct {
	item           StepFlowItem
	transitionsMap map[string][]Transition
}

func newStepFlow(item StepFlowItem) (StepFlow, error) {
	transitions, err := item.Transitions()
	if err != nil {
		return nil, err
	}

	transitionsMap := make(map[string][]Transition)
	for _, t := range transitions {
		transitionsMap[t.Source()] = append(transitionsMap[t.Source()], t)
	}

	return &stepFlow{item: item, transitionsMap: transitionsMap}, nil
}

const ApplyOneMaxIterations = 100

func (sf *stepFlow) Apply(ctx context.Context, oldState []string) ([]string, error) {
	newState := withDefaultValue(oldState, []string{StartCommand(sf.item)})
	var isExclusive bool
	var err error

	for range ApplyOneMaxIterations {
		newState, isExclusive, err = sf.applyOne(ctx, newState)
		if err != nil || isExclusive {
			break
		}
	}

	return newState, err
}

func (sf *stepFlow) applyOne(ctx context.Context, oldState []string) ([]string, bool, error) {
	if sf.IsCompleted(oldState) {
		return oldState, true, nil
	}

	for _, lastEvent := range oldState {
		for _, t := range sf.transitionsMap[lastEvent] {
			isExclusive := t.IsExclusive()
			newState, err := t.Destination(ctx)
			return newState, isExclusive, err
		}
	}

	return nil, true, fmt.Errorf("unhnadled state %s", oldState)
}

func withDefaultValue(value []string, defaultValue []string) []string {
	if value == nil {
		return defaultValue
	}

	return value
}

func (sf *stepFlow) IsCompleted(state []string) bool {
	if len(state) != 1 {
		return false
	}

	return state[0] == CompletedEvent(sf.item)
}

func eventString(item StepFlowItem, event string) string {
	return fmt.Sprintf("%s:%s", event, item.Name())
}

func StartCommand(item StepFlowItem) string {
	return eventString(item, "start")
}

func CompletedEvent(item StepFlowItem) string {
	return eventString(item, "completed")
}

type Transition interface {
	Source() string
	Destination(context.Context) ([]string, error)
	IsExclusive() bool
}

type staticTransition struct {
	source      string
	destination string
}

func (t staticTransition) Source() string {
	return t.source
}

func (t staticTransition) Destination(_ context.Context) ([]string, error) {
	return []string{t.destination}, nil
}

func (t staticTransition) IsExclusive() bool {
	return false
}

type dynamicTransition struct {
	source          string
	destinationFunc func(context.Context) ([]string, error)
}

func (t dynamicTransition) Source() string {
	return t.source
}

func (t dynamicTransition) Destination(ctx context.Context) ([]string, error) {
	return t.destinationFunc(ctx)
}

func (t dynamicTransition) IsExclusive() bool {
	return true
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

		item, err := newNamedItem(namespacedName(namespace, name), maybeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create new named step flow item due to %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func namespacedName(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func newNamedItem(name string, maybeItem any) (StepFlowItem, error) {
	switch maybeItemV := maybeItem.(type) {
	case StepFlowItem:
		return maybeItemV.WithName(name), nil

	case func(context.Context) error:
		return newFuncItem(maybeItemV).WithName(name), nil

	default:
		return nil, fmt.Errorf("type %T is not supported", maybeItemV)
	}
}
