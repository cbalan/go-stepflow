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

	maxIterations := 17

	flow, err := stepflow.NewNamedStepFlow("TestSteps",
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
	for i := range maxIterations {
		t.Logf("[%d] appling stepflow on state %s", i, state)

		ctx := context.WithValue(context.TODO(), exContextKey, &ex)
		state, err = flow.Apply(ctx, state)
		if err != nil {
			t.Fatal(err)
		}

		if flow.IsCompleted(state) {
			break
		}
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

// func TestLoopUntil(t *testing.T) {
// 	startStepA := func(ctx context.Context) error {
// 		return nil
// 	}

// 	stepACompleted := func(ctx context.Context) (bool, error) {
// 		return true, nil
// 	}

// 	stepB := func(ctx context.Context) error {
// 		return nil
// 	}

// 	steps := stepflow.Steps(
// 		startStepA,
// 		stepflow.LoopUntil(stepACompleted),
// 		stepB,
// 	)

// 	expectedIterations := 4

// 	var state []string
// 	for i := range expectedIterations {
// 		t.Logf("[%d] appling stepflow on state %s", i, state)
// 		var err error
// 		state, err = steps.Apply(context.TODO(), state)
// 		if err != nil {
// 			t.Fatal()
// 		}
// 	}

// 	// stepflow should have been completed after the extected number of iterations.
// 	if !stepflow.IsCompleted(state) {
// 		t.Fatalf("Unexpected state %s", state)
// 	}
// }

// func TestConditions(t *testing.T) {

// 	stepA := func(ctx context.Context) error {

// 	}

// 	subStepsA := stepflow.Steps()
// 	subStepsB := stepflow.Steps()

// 	condition := func(context.Context) (bool, error) {
// 		return false, nil
// 	}

// 	steps := stepflow.Steps(
// 		stepA,
// 		stepflow.Case(condition, subStepsA),
// 		stepflow.Case(condition, subStepsB),
// 		stepB,
// 	)

// 	fmt.Println(steps)
// }
