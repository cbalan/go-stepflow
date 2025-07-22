# go-stepflow

**go-stepflow** is a Go library for building and executing durable long-running processes.
It was designed originally to be used in conjunction with kubebuilder/controller-runtime to build Kubernetes controllers, but it has no dependencies to Kubernetes.

## Example
A continuous deployment pipeline step with the following structure may be implemented using the code snippets below.
 * Trigger async actual state collection for a given application in a given environment.
 * Wait for the async operation to complete
 * Compare the desired state with the actual state. If change is required:
   - Create an automated git pull request to the GitOps repository
   - Wait for the automated pull request to be merged automatically.
   - Wait for the GitOps agent(ie. ArgoCD) to successfully uptake the new change. Report back any deployment related events(ie. rolling upgrade details).
   - Trigger post deployment async actual state collections.
   - Wait for the async operation to complete.
   - Validate the post deployment actual state.


### Workflow definition

```go
package demo

import (
	"github.com/cbalan/go-stepflow"
)

func DeployStepFlow() (stepflow.StepFlow, error) {
	return stepflow.NewStepFlow(stepflow.Steps().
		WithName("deploy.v1").
		Do("getPreActualState", getActualState("preDeploy")).
		WaitFor("preActualState", isGetActualStateCompleted("preDeploy")).
		Case("shouldDeploy", shouldDeploy, stepflow.Steps().
			Do("createDeployRequest", createGitOpsVersionBump).
			WaitFor("acceptedDeployRequest", isGitOpsVersionBumpAccepted).
			WaitFor("monitorDeploymentProgress", isVersionLiveSuccessfully).
			Do("getPostActualState", getActualState("postDeploy")).
			WaitFor("postActualState", isGetActualStateCompleted("postDeploy")).
			Do("validatePostDeployActualState", validatePostActuatState)))
}

...

func createGitOpsVersionBump(ctx context.Context) error {
  ... 

  err := gitClient.CreatePullRequest(...)
  ...
  
  return nil
}
```

### Workflow initialization in the main function
```go
  flow, err := DeployStepFlow()
```


### Workflow execution loop
```go
// Prepare execution context
applyCtx, err := setFlowApplyContext(ctx)

// Load previous state
previousState, err := getState(ctx)

// Execute workflow iteration 
newState, err := flow.Apply(applyCtx, previousState)

// Save new state
err := saveState(newState)
```

### Underlying stepflow state machine
```mermaid
stateDiagram-v2
   [*] --> start&colon;deploy.v1
completed&colon;deploy.v1 --> [*]
state deploy.v1 {
state deploy.v1/getPreActualState {
start&colon;deploy.v1/getPreActualState
completed&colon;deploy.v1/getPreActualState
}
state deploy.v1/preActualStateWaitFor {
start&colon;deploy.v1/preActualStateWaitFor
completed&colon;deploy.v1/preActualStateWaitFor
}
state deploy.v1/shouldDeployCase {
state deploy.v1/shouldDeployCase/steps {
state deploy.v1/shouldDeployCase/steps/validatePostDeployActualState {
start&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState
completed&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState
}
state deploy.v1/shouldDeployCase/steps/createDeployRequest {
start&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest
completed&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest
}
state deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor {
start&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor
completed&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor
}
state deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor {
start&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor
completed&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor
}
state deploy.v1/shouldDeployCase/steps/getPostActualState {
start&colon;deploy.v1/shouldDeployCase/steps/getPostActualState
completed&colon;deploy.v1/shouldDeployCase/steps/getPostActualState
}
state deploy.v1/shouldDeployCase/steps/postActualStateWaitFor {
start&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor
completed&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor
}
start&colon;deploy.v1/shouldDeployCase/steps
completed&colon;deploy.v1/shouldDeployCase/steps
}
start&colon;deploy.v1/shouldDeployCase
completed&colon;deploy.v1/shouldDeployCase
}
start&colon;deploy.v1
completed&colon;deploy.v1
}
completed&colon;deploy.v1/getPreActualState --> start&colon;deploy.v1/preActualStateWaitFor: static
start&colon;deploy.v1/preActualStateWaitFor --> start&colon;deploy.v1/preActualStateWaitFor: WaitFor condition is not met
start&colon;deploy.v1/preActualStateWaitFor --> completed&colon;deploy.v1/preActualStateWaitFor: WaitFor condition is met
completed&colon;deploy.v1/preActualStateWaitFor --> start&colon;deploy.v1/shouldDeployCase: static
start&colon;deploy.v1/shouldDeployCase/steps --> start&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest: static
completed&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor: static
completed&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState: static
completed&colon;deploy.v1/shouldDeployCase --> completed&colon;deploy.v1: static
start&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest --> completed&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest: completed
start&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor --> completed&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor: WaitFor condition is met
start&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor: WaitFor condition is not met
start&colon;deploy.v1 --> start&colon;deploy.v1/getPreActualState: static
start&colon;deploy.v1/shouldDeployCase --> start&colon;deploy.v1/shouldDeployCase/steps: Case condition is met
start&colon;deploy.v1/shouldDeployCase --> completed&colon;deploy.v1/shouldDeployCase: Case condition is not met
completed&colon;deploy.v1/shouldDeployCase/steps/createDeployRequest --> start&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor: static
start&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor: WaitFor condition is not met
start&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor --> completed&colon;deploy.v1/shouldDeployCase/steps/acceptedDeployRequestWaitFor: WaitFor condition is met
completed&colon;deploy.v1/shouldDeployCase/steps/getPostActualState --> start&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor: static
start&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor: WaitFor condition is not met
start&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor --> completed&colon;deploy.v1/shouldDeployCase/steps/postActualStateWaitFor: WaitFor condition is met
start&colon;deploy.v1/getPreActualState --> completed&colon;deploy.v1/getPreActualState: completed
completed&colon;deploy.v1/shouldDeployCase/steps --> completed&colon;deploy.v1/shouldDeployCase: static
completed&colon;deploy.v1/shouldDeployCase/steps/monitorDeploymentProgressWaitFor --> start&colon;deploy.v1/shouldDeployCase/steps/getPostActualState: static
start&colon;deploy.v1/shouldDeployCase/steps/getPostActualState --> completed&colon;deploy.v1/shouldDeployCase/steps/getPostActualState: completed
start&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState --> completed&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState: completed
completed&colon;deploy.v1/shouldDeployCase/steps/validatePostDeployActualState --> completed&colon;deploy.v1/shouldDeployCase/steps: static
```