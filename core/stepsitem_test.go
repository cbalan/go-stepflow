package core_test

import (
	"context"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewStepsItem(t *testing.T) {
	// Create a steps item with no child items
	item := core.NewStepsItem("test", []core.StepFlowItem{})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewStepsItem returned nil")
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

	// For empty steps, we should have only one transition (start to completed)
	if len(transitions) != 1 {
		t.Fatalf("Expected 1 transition, got %d", len(transitions))
	}

	// Check transition source and destination
	expectedSource := core.StartCommand(scope)
	expectedDest := core.CompletedEvent(scope)

	if transitions[0].Source().Name() != expectedSource.Name() ||
		transitions[0].Source().Scope().Name() != expectedSource.Scope().Name() {
		t.Fatalf("Expected source %v, got %v", expectedSource, transitions[0].Source())
	}

	destinations, err := transitions[0].Destination(context.Background())
	if err != nil {
		t.Fatalf("Destination returned an error: %v", err)
	}
	if len(destinations) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(destinations))
	}
	if destinations[0].Name() != expectedDest.Name() ||
		destinations[0].Scope().Name() != expectedDest.Scope().Name() {
		t.Fatalf("Expected destination %v, got %v", expectedDest, destinations[0])
	}
}

func TestStepsItem_WithItems(t *testing.T) {
	// Create two function items to add to steps
	item1 := core.NewFuncItem("item1", func(ctx context.Context) error {
		return nil
	})
	item2 := core.NewFuncItem("item2", func(ctx context.Context) error {
		return nil
	})

	// Create a steps item with two child items
	item := core.NewStepsItem("test", []core.StepFlowItem{item1, item2})

	// Get transitions
	_, transitions, err := item.Transitions(nil)
	if err != nil {
		t.Fatalf("Transitions returned an error: %v", err)
	}

	// We should have 5 transitions:
	// 1. Start steps -> Start item1
	// 2. Start item1 -> Completed item1 (from item1)
	// 3. Completed item1 -> Start item2
	// 4. Start item2 -> Completed item2 (from item2)
	// 5. Completed item2 -> Completed steps
	if len(transitions) != 5 {
		t.Fatalf("Expected 5 transitions, got %d", len(transitions))
	}
}

func TestStepsItem_DuplicateNames(t *testing.T) {
	// Create two function items with the same name
	item1 := core.NewFuncItem("duplicate", func(ctx context.Context) error {
		return nil
	})
	item2 := core.NewFuncItem("duplicate", func(ctx context.Context) error {
		return nil
	})

	// Create a steps item with duplicate named items
	item := core.NewStepsItem("test", []core.StepFlowItem{item1, item2})

	// Get transitions - should fail
	_, _, err := item.Transitions(nil)

	// Check that we got an error about duplicate names
	if err == nil {
		t.Fatal("Expected error for duplicate names, got nil")
	}
}

func TestStepsItem_WithParent(t *testing.T) {
	// Create parent scope
	parent := core.NewScope("parent")

	// Create a steps item
	item := core.NewStepsItem("test", []core.StepFlowItem{})

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
