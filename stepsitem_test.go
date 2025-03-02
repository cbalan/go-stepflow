package stepflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cbalan/go-stepflow"
)

func TestSteps(t *testing.T) {
	type contextKey string
	const exContextKey = contextKey("ex")

	stepA := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}
		*ex = append(*ex, "stepA")

		return nil
	}

	stepB := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}
		*ex = append(*ex, "stepB")

		return nil
	}

	stepC := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		*ex = append(*ex, "stepC")

		return nil
	}

	logExchange := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		t.Logf("Current exchange: %s", ex)

		return nil
	}

	flow, err := stepflow.NewStepFlow("TestSteps",
		"stepA", stepA,
		"stepB", stepB,
		"stepC", stepC,
		"stepD", stepflow.Steps(
			"stepA", stepA,
			"stepB", stepB,
			"stepA-bis", stepA),
		"logExchange", logExchange,
	)

	if err != nil {
		t.Fatal(err)
	}

	// Execute the stepflow
	var ex []string
	var state []string

	expectedIterations := 17
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

	expectedExString := "[stepA stepB stepC stepA stepB stepA]"
	if fmt.Sprintf("%s", ex) != expectedExString {
		t.Fatalf("Unexpected exchange. Expected: %s, Actual: %s", expectedExString, ex)
	}
}
