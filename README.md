# go-stepflow

**go-stepflow** is a Go library for building and executing stateful, resumable workflows using a fluent API.

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

func main() {
    // Define workflow
    flow, err := stepflow.NewStepFlow(stepflow.Steps().
        Do("step1", func(ctx context.Context) error {
            fmt.Println("Executing step 1")
            return nil
        }).
        WaitFor("condition", func(ctx context.Context) (bool, error) {
            // Return true when ready to proceed
            return true, nil
        }).
        Do("step2", func(ctx context.Context) error {
            fmt.Println("Executing step 2")
            return nil
        }))

    if err != nil {
        panic(err)
    }

    // Execute workflow
    var state []string // Could be loaded from persistent storage
    for !flow.IsCompleted(state) {
        state, err = flow.Apply(context.Background(), state)
        if err != nil {
            panic(err)
        }
        // Persist state here if needed
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