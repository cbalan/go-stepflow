package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewCaseItem(t *testing.T) {
	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return nil
	})

	// Create a case item
	item := core.NewCaseItem("test", child, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewCaseItem returned nil")
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

	// For a case with a child, we should have 3 transitions:
	// 1. Start case -> Start child or Completed case (from condition)
	// 2. Start child -> Completed child (from child)
	// 3. Completed child -> Completed case
	if len(transitions) != 3 {
		t.Fatalf("Expected 3 transitions, got %d", len(transitions))
	}

	// Check transition source and destination for the conditional transition
	expectedSource := core.StartCommand(scope)

	// Find the transition from start
	var startTransition core.Transition
	for _, t := range transitions {
		if t.Source().Name() == expectedSource.Name() &&
			t.Source().Scope().Name() == expectedSource.Scope().Name() {
			startTransition = t
			break
		}
	}

	if startTransition == nil {
		t.Fatal("Could not find transition from start event")
	}

	// Check that the transition's possible destinations are correct
	possibleDests := startTransition.PossibleDestinations()
	if len(possibleDests) != 2 {
		t.Fatalf("Expected 2 possible destinations, got %d", len(possibleDests))
	}

	// The possibilities should be StartCommand(childScope) and CompletedEvent(scope)
	foundChildStart := false
	foundCaseCompleted := false
	for _, pd := range possibleDests {
		if pd.Event().Name() == "start" && pd.Event().Scope().Name() != scope.Name() {
			foundChildStart = true
		}
		if pd.Event().Name() == "completed" && pd.Event().Scope().Name() == scope.Name() {
			foundCaseCompleted = true
		}
	}

	if !foundChildStart {
		t.Fatal("Expected child start event in possible destinations")
	}
	if !foundCaseCompleted {
		t.Fatal("Expected case completed event in possible destinations")
	}
}

func TestCaseItem_ConditionTrue(t *testing.T) {
	// Keep track of whether the child function was called
	childCalled := false

	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		childCalled = true
		return nil
	})

	// Create a case item with condition that returns true
	item := core.NewCaseItem("test", child, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Create a step flow with the case item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	var state []string
	var errApply error

	expectedIterations := 3
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

	// Check that the child function was called
	if !childCalled {
		t.Fatal("Child function should have been called when condition is true")
	}
}

func TestCaseItem_ConditionFalse(t *testing.T) {
	// Keep track of whether the child function was called
	childCalled := false

	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		childCalled = true
		return nil
	})

	// Create a case item with condition that returns false
	item := core.NewCaseItem("test", child, func(ctx context.Context) (bool, error) {
		return false, nil
	})

	// Create a step flow with the case item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	state, err := sf.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("Apply returned an error: %v", err)
	}

	// Check that the child function was not called
	if childCalled {
		t.Fatal("Child function should not have been called when condition is false")
	}

	// Check that the flow is completed
	if !sf.IsCompleted(state) {
		t.Fatal("Flow should be completed even when condition is false")
	}
}

func TestCaseItem_Error(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Create a child item
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return nil
	})

	// Create a case item with condition that returns an error
	item := core.NewCaseItem("test", child, func(ctx context.Context) (bool, error) {
		return false, expectedErr
	})

	// Create a step flow with the case item
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
