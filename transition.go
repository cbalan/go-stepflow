package stepflow

import "context"

type Transition interface {
	Source() string
	Destination(context.Context) ([]string, error)
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
