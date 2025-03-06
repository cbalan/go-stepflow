package stepflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cbalan/go-stepflow"
)

func TestStepFlowApply(t *testing.T) {
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

	flow, err := stepflow.NewStepFlow("TestStepFlowApply",
		"stepA", stepA,
		"stepB", stepB,
		"stepC", stepC,
	)

	if err != nil {
		t.Fatal(err)
	}

	// Execute the stepflow
	var ex []string
	var state []string

	expectedIterations := 4
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

	expectedExString := "[stepA stepB stepC]"
	if fmt.Sprintf("%s", ex) != expectedExString {
		t.Fatalf("Unexpected exchange. Expected: %s, Actual: %s", expectedExString, ex)
	}
}
