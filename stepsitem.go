package stepflow

type ItemsProvider interface {
	Items(namespace string) ([]StepFlowItem, error)
}

type stepsItem struct {
	name          string
	itemsProvider ItemsProvider
}

func newStepsItem(itemsProvider ItemsProvider) StepFlowItem {
	return &stepsItem{itemsProvider: itemsProvider}
}

func (si *stepsItem) Name() string {
	return si.name
}

func (si *stepsItem) WithName(name string) StepFlowItem {
	return &stepsItem{name: name, itemsProvider: si.itemsProvider}
}

func (si *stepsItem) Transitions() ([]Transition, error) {
	items, err := si.itemsProvider.Items(si.name)
	if err != nil {
		return nil, err
	}

	lastEvent := StartCommand(si)

	var transitions []Transition
	for _, item := range items {
		transitions = append(transitions, staticTransition{source: lastEvent, destination: StartCommand(item)})

		itemTransitions, err := item.Transitions()
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, itemTransitions...)

		lastEvent = CompletedEvent(item)
	}

	transitions = append(transitions, staticTransition{source: lastEvent, destination: CompletedEvent(si)})

	return transitions, nil
}
