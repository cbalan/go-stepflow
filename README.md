# go-stepflow

**go-stepflow** is a Go library for building and executing resumable workflows using a fluent API.

It is designed for building systems that need to manage complex multi-step processes over time, such as:
- Continuous deployment pipelines
- Business process orchestration
- Multi-stage data processing
- Long-running external service integrations

## Key Features
- **Fluent, Declarative API** - Build workflows by chaining intuitive method calls.
- **State Machine Foundation** -  Execution model based on event transitions.
- **Persistence & Resumability** - Save workflow state after each step for fault tolerance.
- **Composable Step Patterns** - Combine and nest steps to create complex workflows.
- **Built-in Control Flow** - Conditional execution, loops, and error handling.

> **Durable workflows**
> This library only provides the means to pause a workflow and serialize its state after each step, based on the provided definition.
> To implement durable workflows, this library must be paired with systems that provide persistent storage and distributed locks.

## Installation
```bash
go get github.com/cbalan/go-stepflow
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/cbalan/go-stepflow"
)

func step1(ctx context.Context) error {
	fmt.Println("Executing step 1")
	return nil
}

func aCondition(ctx context.Context) (bool, error) {
	fmt.Println("Return true when ready to proceed.")
	return true, nil
}

func step2(ctx context.Context) error {
	fmt.Println("Executing step 2")
	return nil
}

func quickStartStepFlow() (stepflow.StepFlow, error) {
	return stepflow.NewStepFlow(stepflow.Named("quickstart.v1").
		Do("step1", step1).
		WaitFor("aCondition", aCondition).
		Do("step2", step2))
}

func main() {
	// Workflow definition.
	flow, err := quickStartStepFlow()
	if err != nil {
		panic(err)
	}

	// Workflow execution.
	var state []string
	for !flow.IsCompleted(state) {
		// Could load state from persistent storage.
		
		// Apply workflow on the old state.
		state, err = flow.Apply(context.Background(), state)
		if err != nil {
			panic(err)
		}
		
		// Could save state to persistent storage.
	}
}
```

## Core Building Blocks
go-stepflow provides several key components for building workflows:

### Step Types
- **`Do(name, func)`** - Execute a function
- **`WaitFor(name, conditionFunc)`** - Execute conditionFunc in a loop until the wait condition is met and the workflow can proceed to the next step.
- **`Steps(name, steps)`** - Group multiple steps together.
- **`Case(name, conditionFunc, steps)`** - Conditional execution.
- **`Retry(name, errorHandlerFunc, steps)`** - Error handling with retry logic.
- **`LoopUntil(name, conditionFunc, steps)`** - Repeat steps until condition is met.

### Example Workflow
```go
workflow, err := stepflow.NewStepFlow(stepflow.Steps()
    Do("prepare", prepareEnvironment).
    WaitFor("ready", isEnvironmentReady).
    Case("shouldDeploy", shouldDeployNewVersion, stepflow.Steps().
        Do("deploy", deployNewVersion).
        WaitFor("deployed", isDeploymentComplete).
        Do("validate", validateDeployment)).
    Do("cleanup", cleanupResources))
```

Please visit [go-stepflow-examples](https://github.com/cbalan/go-stepflow-examples) for additional examples.