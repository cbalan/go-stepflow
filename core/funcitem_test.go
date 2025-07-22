package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewFuncItem(t *testing.T) {
	// Create a function item
	called := false
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		called = true
		return nil
	})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewFuncItem returned nil")
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

	// Check the function is called when the transition is triggered
	destinations, err := transitions[0].Destination(context.Background())
	if err != nil {
		t.Fatalf("Destination returned an error: %v", err)
	}

	// Check the function was called
	if !called {
		t.Fatal("Function was not called")
	}

	// Check destinations
	if len(destinations) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(destinations))
	}

	// Check destination event
	expectedDest := core.CompletedEvent(scope)
	if destinations[0].Name() != expectedDest.Name() ||
		destinations[0].Scope().Name() != expectedDest.Scope().Name() {
		t.Fatalf("Expected destination %v, got %v", expectedDest, destinations[0])
	}
}

func TestFuncItem_Error(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Create a function item that returns an error
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		return expectedErr
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

func TestFuncItem_WithParent(t *testing.T) {
	// Create parent scope
	parent := core.NewScope("parent")

	// Create a function item
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		return nil
	})

	// Get transitions
	scope, _, err := item.Transitions(parent)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// Check the scope
	expectedName := "parent/test"
	if scope.Name() != expectedName {
		t.Fatalf("Expected scope name '%s', got '%s'", expectedName, scope.Name())
	}

	// Check the parent
	if scope.Parent() != parent {
		t.Fatalf("Expected parent %v, got %v", parent, scope.Parent())
	}
}
