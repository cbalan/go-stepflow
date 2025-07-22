package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewLoopUntilItem(t *testing.T) {
	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return nil
	})

	// Create a loop until item
	item := core.NewLoopUntilItem("test", child, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewLoopUntilItem returned nil")
	}

	// Get transitions
	scope, transitions, err := item.Transitions(nil)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// Check the scope
	if scope == nil {
		t.Fatal("Transitions returned nil scope")
	}
	if scope.Name() != "test" {
		t.Fatalf("Expected scope name 'test', got '%s'", scope.Name())
	}

	// For a loop with a child, we should have 3 transitions:
	// 1. Start loop -> Start child
	// 2. Start child -> Completed child (from child)
	// 3. Completed child -> Start child or Completed loop (from condition)
	if len(transitions) != 3 {
		t.Fatalf("Expected 3 transitions, got %d", len(transitions))
	}

	// Find the transition from completed child
	var conditionTransition core.Transition
	for _, t := range transitions {
		if t.Source().Name() == "completed" && t.Source().Scope().Name() != scope.Name() {
			conditionTransition = t
			break
		}
	}

	if conditionTransition == nil {
		t.Fatal("Could not find transition from completed child event")
	}

	// Check that the transition's possible destinations are correct
	possibleDests := conditionTransition.PossibleDestinations()
	if len(possibleDests) != 2 {
		t.Fatalf("Expected 2 possible destinations, got %d", len(possibleDests))
	}

	// The possibilities should be StartCommand(childScope) and CompletedEvent(scope)
	foundChildStart := false
	foundLoopCompleted := false
	for _, pd := range possibleDests {
		if pd.Event().Name() == "start" && pd.Event().Scope().Name() != scope.Name() {
			foundChildStart = true
		}
		if pd.Event().Name() == "completed" && pd.Event().Scope().Name() == scope.Name() {
			foundLoopCompleted = true
		}
	}

	if !foundChildStart {
		t.Fatal("Expected child start event in possible destinations")
	}
	if !foundLoopCompleted {
		t.Fatal("Expected loop completed event in possible destinations")
	}
}

func TestLoopUntilItem_ConditionTrue(t *testing.T) {
	// Create a child item
	callCount := 0
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		callCount++
		return nil
	})

	// Create a loop until item with condition that returns true immediately
	item := core.NewLoopUntilItem("test", child, func(ctx context.Context) (bool, error) {
		return true, nil // Condition met on first check
	})

	// Create a step flow with the loop item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	var state []string
	var errApply error

	expectedIterations := 2
	for range expectedIterations {
		state, errApply = sf.Apply(context.Background(), state)
		if errApply != nil {
			t.Fatalf("Apply returned an error: %v", err)
		}
	}

	// stepflow should have been completed after the expected number of iterations.
	if !sf.IsCompleted(state) {
		t.Fatalf("Unexpected state %s", state)
	}

	// Check that the child function was called exactly once
	if callCount != 1 {
		t.Fatalf("Expected child function to be called 1 time, got %d", callCount)
	}
}

func TestLoopUntilItem_MultipleIterations(t *testing.T) {
	// Create a child item
	callCount := 0
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		callCount++
		return nil
	})

	// Create a loop until item with condition that returns true after 3 iterations
	item := core.NewLoopUntilItem("test", child, func(ctx context.Context) (bool, error) {
		return callCount >= 3, nil // Condition met after 3 iterations
	})

	// Create a step flow with the loop item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	var state []string
	var errApply error

	expectedIterations := 6
	for range expectedIterations {
		state, errApply = sf.Apply(context.Background(), state)
		if errApply != nil {
			t.Fatalf("Apply returned an error: %v", err)
		}
	}

	// stepflow should have been completed after the expected number of iterations.
	if !sf.IsCompleted(state) {
		t.Fatalf("Unexpected state %s", state)
	}

	// Check that the child function was called exactly 3 times
	if callCount != 3 {
		t.Fatalf("Expected child function to be called 3 times, got %d", callCount)
	}
}

func TestLoopUntilItem_ChildError(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Create a child item that will fail
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return expectedErr
	})

	// Create a loop until item
	item := core.NewLoopUntilItem("test", child, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Create a step flow with the loop item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	_, err = sf.Apply(context.Background(), nil)

	// Check the error
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestLoopUntilItem_ConditionError(t *testing.T) {
	// Create an error
	expectedErr := errors.New("condition error")

	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return nil
	})

	// Create a loop until item with condition that returns an error
	item := core.NewLoopUntilItem("test", child, func(ctx context.Context) (bool, error) {
		return false, expectedErr
	})

	// Create a step flow with the loop item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	var state []string
	var errApply error

	expectedIterations := 2
	for range expectedIterations {
		state, errApply = sf.Apply(context.Background(), state)
		if errApply != nil {
			break
		}
	}

	// Check the error
	if !errors.Is(errApply, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}
