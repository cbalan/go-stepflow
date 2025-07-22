package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cbalan/go-stepflow/core"
)

// TestComplexWorkflow tests a complex workflow that uses multiple components together
func TestComplexWorkflow(t *testing.T) {
	// Track execution order
	executionOrder := []string{}

	// Create a function item
	funcItem := core.NewFuncItem("func", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "func")
		return nil
	})

	// Create a case item that will execute
	caseItem := core.NewCaseItem("case", funcItem, func(ctx context.Context) (bool, error) {
		executionOrder = append(executionOrder, "case-condition")
		return true, nil
	})

	// Create a loop that runs twice
	iterationCount := 0
	loopItem := core.NewLoopUntilItem("loop", funcItem, func(ctx context.Context) (bool, error) {
		executionOrder = append(executionOrder, "loop-condition")
		iterationCount++
		return iterationCount >= 2, nil
	})

	// Create a steps item that executes the case and loop in sequence
	stepsItem := core.NewStepsItem("steps", []core.StepFlowItem{caseItem, loopItem})

	// Create a step flow with the steps item
	sf, err := core.NewStepFlow(stepsItem)
	if err != nil {
		t.Fatalf("NewStepFlow returned an error: %v", err)
	}

	// Apply the step flow
	var state []string
	var errApply error

	expectedIterations := 7
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

	// Expected execution order:
	// 1. case-condition
	// 2. func (from case)
	// 3. func (first loop iteration)
	// 4. loop-condition (first check)
	// 5. func (second loop iteration)
	// 6. loop-condition (second check, returns true)
	expectedOrder := []string{
		"case-condition",
		"func", // from case
		"func", // first loop iteration
		"loop-condition",
		"func", // second loop iteration
		"loop-condition",
	}

	// Check the execution order
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d executions, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, step := range expectedOrder {
		if executionOrder[i] != step {
			t.Fatalf("Expected step %d to be %s, got %s", i, step, executionOrder[i])
		}
	}
}

// TestErrorHandling tests how errors propagate through the workflow
func TestErrorHandling(t *testing.T) {
	// Create an error
	expectedErr := errors.New("test error")

	// Track retries
	retryCount := 0

	// Create a function item that always fails
	funcItem := core.NewFuncItem("func", func(ctx context.Context) error {
		return expectedErr
	})

	// Create a retry item that retries 3 times then gives up
	retryItem := core.NewRetryItem(funcItem, func(ctx context.Context, err error) (bool, error) {
		retryCount++
		return retryCount < 3, nil // Retry 3 times then stop
	})

	// Create a step flow with the retry item
	sf, err := core.NewStepFlow(retryItem)
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
			break
		}
	}

	// Check the error
	if !errors.Is(errApply, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}

	// Check that it retried the expected number of times
	if retryCount != 3 {
		t.Fatalf("Expected 3 retries, got %d", retryCount)
	}
}

// TestWaitForWithCase tests WaitFor and Case items working together
func TestWaitForWithCase(t *testing.T) {
	// Track execution
	waitCount := 0
	funcCalled := false

	// Create a wait for item that waits 3 times
	waitForItem := core.NewWaitForItem("wait", func(ctx context.Context) (bool, error) {
		waitCount++
		return waitCount >= 3, nil // Complete after 3 checks
	})

	// Create a function item
	funcItem := core.NewFuncItem("func", func(ctx context.Context) error {
		funcCalled = true
		return nil
	})

	// Create a case item that executes the function only if wait completed
	caseItem := core.NewCaseItem("case", funcItem, func(ctx context.Context) (bool, error) {
		return waitCount >= 3, nil
	})

	// Create a steps item that executes the wait and case in sequence
	stepsItem := core.NewStepsItem("steps", []core.StepFlowItem{waitForItem, caseItem})

	// Create a step flow with the steps item
	sf, err := core.NewStepFlow(stepsItem)
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

	// Check that the wait for executed 3 times
	if waitCount != 3 {
		t.Fatalf("Expected wait to execute 3 times, got %d", waitCount)
	}

	// Check that the function was called
	if !funcCalled {
		t.Fatal("Function should have been called")
	}
}
