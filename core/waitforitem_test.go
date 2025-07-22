package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewWaitForItem(t *testing.T) {
	// Create a wait for item
	item := core.NewWaitForItem("test", func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewWaitForItem returned nil")
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

	// Check transitions
	if len(transitions) != 1 {
		t.Fatalf("Expected 1 transition, got %d", len(transitions))
	}

	// Check transition source
	expectedSource := core.StartCommand(scope)
	if transitions[0].Source().Name() != expectedSource.Name() ||
		transitions[0].Source().Scope().Name() != expectedSource.Scope().Name() {
		t.Fatalf("Expected source %v, got %v", expectedSource, transitions[0].Source())
	}

	// Check that the transition's possible destinations are correct
	possibleDests := transitions[0].PossibleDestinations()
	if len(possibleDests) != 2 {
		t.Fatalf("Expected 2 possible destinations, got %d", len(possibleDests))
	}

	// The possibilities should be StartCommand(scope) and CompletedEvent(scope)
	foundStart := false
	foundCompleted := false
	for _, pd := range possibleDests {
		if pd.Event().Name() == "start" && pd.Event().Scope().Name() == scope.Name() {
			foundStart = true
		}
		if pd.Event().Name() == "completed" && pd.Event().Scope().Name() == scope.Name() {
			foundCompleted = true
		}
	}

	if !foundStart {
		t.Fatal("Expected start event in possible destinations")
	}
	if !foundCompleted {
		t.Fatal("Expected completed event in possible destinations")
	}
}

func TestWaitForItem_ConditionMet(t *testing.T) {
	// Create a wait for item with condition that returns true (met)
	item := core.NewWaitForItem("test", func(ctx context.Context) (bool, error) {
		return true, nil
	})

	// Get transitions
	scope, transitions, err := item.Transitions(nil)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// Execute the transition
	destinations, err := transitions[0].Destination(context.Background())
	if err != nil {
		t.Fatalf("Destination returned an error: %v", err)
	}

	// Check destinations - should be CompletedEvent since condition is met
	if len(destinations) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(destinations))
	}

	expectedDest := core.CompletedEvent(scope)
	if destinations[0].Name() != expectedDest.Name() ||
		destinations[0].Scope().Name() != expectedDest.Scope().Name() {
		t.Fatalf("Expected destination %v, got %v", expectedDest, destinations[0])
	}
}

func TestWaitForItem_ConditionNotMet(t *testing.T) {
	// Create a wait for item with condition that returns false (not met)
	item := core.NewWaitForItem("test", func(ctx context.Context) (bool, error) {
		return false, nil
	})

	// Get transitions
	scope, transitions, err := item.Transitions(nil)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// Execute the transition
	destinations, err := transitions[0].Destination(context.Background())
	if err != nil {
		t.Fatalf("Destination returned an error: %v", err)
	}

	// Check destinations - should be StartCommand since condition is not met (loop back)
	if len(destinations) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(destinations))
	}

	expectedDest := core.StartCommand(scope)
	if destinations[0].Name() != expectedDest.Name() ||
		destinations[0].Scope().Name() != expectedDest.Scope().Name() {
		t.Fatalf("Expected destination %v, got %v", expectedDest, destinations[0])
	}
}

func TestWaitForItem_Error(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Create a wait for item with condition that returns an error
	item := core.NewWaitForItem("test", func(ctx context.Context) (bool, error) {
		return false, expectedErr
	})

	// Get transitions
	_, transitions, err := item.Transitions(nil)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// Execute the transition
	_, err = transitions[0].Destination(context.Background())

	// Check the error
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}
