package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

func TestNewRetryItem(t *testing.T) {
	// Create a child item that will fail
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return errors.New("test error")
	})

	// Create a retry item
	item := core.NewRetryItem(child, func(ctx context.Context, err error) (bool, error) {
		return true, nil // Always retry
	})

	// Check that the item is not nil
	if item == nil {
		t.Fatal("NewRetryItem returned nil")
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

	// Check transitions - the retry item should wrap the child's transitions
	if len(transitions) != 1 {
		t.Fatalf("Expected 1 transition, got %d", len(transitions))
	}

	// Check transition source
	expectedSource := core.StartCommand(scope)
	if transitions[0].Source().Name() != expectedSource.Name() ||
		transitions[0].Source().Scope().Name() != expectedSource.Scope().Name() {
		t.Fatalf("Expected source %v, got %v", expectedSource, transitions[0].Source())
	}

	// Check that the transition's possible destinations include retry
	possibleDests := transitions[0].PossibleDestinations()
	foundRetry := false
	for _, pd := range possibleDests {
		if pd.Reason() == "retry" {
			foundRetry = true
			break
		}
	}

	if !foundRetry {
		t.Fatal("Expected retry in possible destinations")
	}
}

func TestRetriableTransition_Success(t *testing.T) {
	// Track calls
	callCount := 0

	// Create a child item that will succeed on the second attempt
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return errors.New("first attempt error")
		}
		return nil
	})

	// Create a retry item
	item := core.NewRetryItem(child, func(ctx context.Context, err error) (bool, error) {
		return true, nil // Always retry
	})

	// Create a step flow with the retry item
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

	// Check that the function was called twice
	if callCount != 2 {
		t.Fatalf("Expected function to be called 2 times, got %d", callCount)
	}
}

func TestRetriableTransition_NoRetry(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Create a child item that will fail
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return expectedErr
	})

	// Create a retry item with handler that doesn't retry
	item := core.NewRetryItem(child, func(ctx context.Context, err error) (bool, error) {
		return false, nil // Don't retry
	})

	// Create a step flow with the retry item
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

func TestRetriableTransition_ErrorHandlerError(t *testing.T) {
	// Create errors
	funcErr := errors.New("function error")
	handlerErr := errors.New("handler error")

	// Create a child item that will fail
	child := core.NewFuncItem("child", func(ctx context.Context) error {
		return funcErr
	})

	// Create a retry item with handler that also fails
	item := core.NewRetryItem(child, func(ctx context.Context, err error) (bool, error) {
		return false, handlerErr // Error in the handler
	})

	// Create a step flow with the retry item
	sf, err := core.NewStepFlow(item)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	_, err = sf.Apply(context.Background(), nil)

	// Check the error - should be the handler error, not the function error
	if err != handlerErr {
		t.Fatalf("Expected error %v, got %v", handlerErr, err)
	}
}
