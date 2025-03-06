package stepflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cbalan/go-stepflow"
)

func TestLoopUntil(t *testing.T) {
	type contextKey string
	const exContextKey = contextKey("ex")

	addA := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}
		*ex = append(*ex, "A")

		return nil
	}

	exLenIsAcceptable := func(ctx context.Context) (bool, error) {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return false, fmt.Errorf("failed to get exchange from context")
		}

		return len(*ex) > 10, nil
	}

	logExchange := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		t.Logf("Current exchange: %s", ex)

		return nil
	}

	flow, err := stepflow.NewStepFlow("TestLoopUntil",
		"addA", addA,
		"growEx", stepflow.LoopUntil(exLenIsAcceptable,
			"addA", addA,
			"logExchange", logExchange,
		),
		"logExchange", logExchange,
	)
	if err != nil {
		t.Fatal(err)
	}

	var ex []string
	var state []string

	expectedIterations := 33
	for i := range expectedIterations {
		t.Logf("[%d] Applying stepflow on state %s", i, state)

		ctx := context.WithValue(context.TODO(), exContextKey, &ex)
		state, err = flow.Apply(ctx, state)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("[%d] Stepflow new state: %s", i, state)
	}

	// stepflow should have been completed after the extected number of iterations.
	if !flow.IsCompleted(state) {
		t.Fatalf("Unexpected state %s", state)
	}
}
