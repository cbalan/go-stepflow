package stepflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cbalan/go-stepflow"
)

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

	flow, err := stepflow.NewStepFlow("TestCase",
		"beforeTrueCase", doLog("beforeTrueCase"),
		"ifTrue", stepflow.Case(alwaysTrue,
			"stepA", doLog("stepA"),
			"stepB", doLog("stepB"),
		),
		"ifFalse", stepflow.Case(alwaysFalse,
			"ifFalseBranch", doLog("this will not be executed"),
		),
		"afterIfFalse", doLog("afterIfFalse"),
	)
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
