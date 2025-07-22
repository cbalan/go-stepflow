package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewStepFlow(t *testing.T) {
	// Create a simple step flow item
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		return nil
	})

	// Create a step flow
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Check that the step flow is not nil
	if sf == nil {
		t.Fatal("NewStepFlow returned nil")
	}
}

func TestStepFlow_Apply(t *testing.T) {
	// Keep track of executed steps
	executed := make(map[string]bool)

	// Create a step flow with a single function item
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		executed["test"] = true
		return nil
	})

	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	state, err := sf.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("Apply returned an error: %v", err)
	}

	// Check that the function was executed
	if !executed["test"] {
		t.Fatal("Function was not executed")
	}

	// Check that the state is not empty
	if len(state) == 0 {
		t.Fatal("Apply returned an empty state")
	}

	// Check that the step flow is completed
	if !sf.IsCompleted(state) {
		t.Fatal("Step flow should be completed")
	}
}

func TestStepFlow_Apply_Error(t *testing.T) {
	// Create a step flow with a function that returns an error
	expectedErr := errors.New("test error")
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		return expectedErr
	})

	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	_, err = sf.Apply(context.Background(), nil)

	// Check that the error was returned
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestStepFlow_IsCompleted(t *testing.T) {
	// Create a step flow with a single function item
	item := core.NewFuncItem("test", func(ctx context.Context) error {
		return nil
	})

	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	state, err := sf.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("Apply returned an error: %v", err)
	}

	// Check that the step flow is completed
	if !sf.IsCompleted(state) {
		t.Fatal("Step flow should be completed")
	}

	// Check that an empty state is not completed
	if sf.IsCompleted([]string{}) {
		t.Fatal("Empty state should not be completed")
	}

	// Check that a different state is not completed
	if sf.IsCompleted([]string{"wrong"}) {
		t.Fatal("Wrong state should not be completed")
	}
}

func TestNewScope(t *testing.T) {
	// Create a new scope
	scope := core.NewScope("test")

	// Check that the scope is not nil
	if scope == nil {
		t.Fatal("NewScope returned nil")
	}

	// Check that the name is correct
	if scope.Name() != "test" {
		t.Fatalf("Expected name 'test', got '%s'", scope.Name())
	}

	// Check that the parent is nil
	if scope.Parent() != nil {
		t.Fatalf("Expected nil parent, got %v", scope.Parent())
	}
}

func TestWithParent(t *testing.T) {
	// Create parent and child scopes
	parent := core.NewScope("parent")
	child := core.NewScope("child")

	// Create a new scope with parent
	scope := core.WithParent(child, parent)

	// Check that the scope is not nil
	if scope == nil {
		t.Fatal("WithParent returned nil")
	}

	// Check that the name is correctly combined
	expectedName := "parent/child"
	if scope.Name() != expectedName {
		t.Fatalf("Expected name '%s', got '%s'", expectedName, scope.Name())
	}

	// Check that the parent is set
	if scope.Parent() != parent {
		t.Fatalf("Expected parent %v, got %v", parent, scope.Parent())
	}
}

func TestWithParent_NilParent(t *testing.T) {
	// Create a child scope
	child := core.NewScope("child")

	// Create a new scope with nil parent
	scope := core.WithParent(child, nil)

	// Check that the scope is the same as the child
	if scope != child {
		t.Fatalf("Expected scope to be the same as child, got %v", scope)
	}
}
