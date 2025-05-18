package stepflow_test

import (
	"context"
	"fmt"
	"github.com/cbalan/go-stepflow"
	"testing"
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

	stepD := stepflow.Steps().
		Do("stepA", stepA).
		Do("stepB", stepB).
		Do("stepA-bis", stepA)

	logExchange := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		t.Logf("Current exchange: %s", ex)

		return nil
	}

	flow, err := stepflow.NewStepFlow("TestSteps", stepflow.Steps().
		Do("stepA", stepA).
		Do("stepB", stepB).
		Do("stepC", stepC).
		Steps("stepD", stepD).
		Do("logExchange", logExchange))
	if err != nil {
		t.Fatal(err)
	}

	// Execute the stepflow
	var ex []string
	var state []string

	expectedIterations := 8
	for i := range expectedIterations {
		t.Logf("[%d] Applying stepflow on state %s", i, state)

		ctx := context.WithValue(context.TODO(), exContextKey, &ex)
		state, err = flow.Apply(ctx, state)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("[%d] Stepflow new state: %s", i, state)
	}

	// stepflow should have been completed after the expected number of iterations.
	if !flow.IsCompleted(state) {
		t.Fatalf("Unexpected state %s", state)
	}

	expectedExString := "[stepA stepB stepC stepA stepB stepA]"
	if fmt.Sprintf("%s", ex) != expectedExString {
		t.Fatalf("Unexpected exchange. Expected: %s, Actual: %s", expectedExString, ex)
	}
}

func TestWaitFor(t *testing.T) {
	type contextKey string
	const exContextKey = contextKey("ex")

	exLenIsAcceptable := func(ctx context.Context) (bool, error) {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return false, fmt.Errorf("failed to get exchange from context")
		}

		*ex = append(*ex, "A")

		return len(*ex) > 3, nil
	}

	logExchange := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		t.Logf("Current exchange: %s", ex)

		return nil
	}

	flow, err := stepflow.NewStepFlow("TestWaitFor", stepflow.Steps().
		WaitFor("growEx", exLenIsAcceptable).
		Do("logExchange", logExchange))
	if err != nil {
		t.Fatal(err)
	}

	var ex []string
	var state []string

	expectedIterations := 6
	for i := range expectedIterations {
		t.Logf("[%d] Applying stepflow on state %s", i, state)

		ctx := context.WithValue(context.TODO(), exContextKey, &ex)
		state, err = flow.Apply(ctx, state)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("[%d] Stepflow new state: %s", i, state)
	}

	// stepflow should have been completed after the expected number of iterations.
	if !flow.IsCompleted(state) {
		t.Fatalf("Unexpected state %s", state)
	}

	expectedExString := "[A A A A]"
	if fmt.Sprintf("%s", ex) != expectedExString {
		t.Fatalf("Unexpected exchange. Expected: %s, Actual: %s", expectedExString, ex)
	}
}

func TestRetry(t *testing.T) {
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

	logErrorAndRetry := func(ctx context.Context, err error) (bool, error) {
		t.Logf("handling error %s", err)
		return true, nil
	}

	returnErrorSometimes := func(ctx context.Context) error {
		ex, ok := ctx.Value(exContextKey).(*[]string)
		if !ok {
			return fmt.Errorf("failed to get exchange from context")
		}

		if len(*ex) < 3 {
			return fmt.Errorf("error")
		}

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

	flow, err := stepflow.NewStepFlow("TestRetry", stepflow.Steps().
		Do("logInitialExchange", logExchange).
		Retry("growEx", logErrorAndRetry, stepflow.Steps().
			Do("addA", addA).
			Do("stepWithError", returnErrorSometimes).
			Do("logExchange", logExchange)).
		Do("logExchange", logExchange))
	if err != nil {
		t.Fatal(err)
	}

	var ex []string
	var state []string

	expectedIterations := 10
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

	flow, err := stepflow.NewStepFlow("TestLoopUntil", stepflow.Steps().
		Do("addA", addA).
		LoopUntil("growEx", exLenIsAcceptable, stepflow.Steps().
			Do("addA", addA).
			Do("logExchange", logExchange)).
		Do("logExchange", logExchange))
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

func TestCase(t *testing.T) {
	type contextKey string
	const exContextKey = contextKey("ex")

	doLog := func(message string) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			ex, ok := ctx.Value(exContextKey).(*[]string)
			if !ok {
				return fmt.Errorf("failed to get exchange from context")
			}
			*ex = append(*ex, message)

			t.Log(message)
			return nil
		}
	}

	alwaysTrue := func(ctx context.Context) (bool, error) {
		return true, nil
	}

	alwaysFalse := func(ctx context.Context) (bool, error) {
		return false, nil
	}

	flow, err := stepflow.NewStepFlow("TestCase", stepflow.Steps().
		Do("beforeTrueCase", doLog("beforeTrueCase")).
		Case("ifTrue", alwaysTrue, stepflow.Steps().
			Do("stepA", doLog("stepA")).
			Do("stepB", doLog("stepB"))).
		Case("ifFalse", alwaysFalse, stepflow.Steps().
			Do("ifFalseBranch", doLog("this will not be executed"))).
		Do("afterIfFalse", doLog("afterIfFalse")))
	if err != nil {
		t.Fatal(err)
	}

	var ex []string
	var state []string

	expectedIterations := 7
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

	expectedExString := "[beforeTrueCase stepA stepB afterIfFalse]"
	if fmt.Sprintf("%s", ex) != expectedExString {
		t.Fatalf("Unexpected exchange. Expected: %s, Actual: %s", expectedExString, ex)
	}
}
