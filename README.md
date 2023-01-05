# Stub Generation for Temporal Activities

> Beware, that this code is currently PoC quality, so expect some issues when using it - please report them!
>
> But this shouldn't stop you from using it, as the _generated code_ is perfectly fine and easy to inspect.

When calling Temporal activities in Go you lose type safety, and it's a bit of boilerplate code.

But imagine instead of this

```go
func BoilerplateWorkflow(ctx workflow.Context) (r *WfResult, err error) {
	var a Activities
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	f := workflow.ExecuteActivity(ctx, a.DoSomething,  workflow.GetInfo(ctx).WorkflowExecution.ID, "maybe a string?")
	var result string
	err = f.Get(ctx, &result)
	if err != nil {
		return nil, err
	}
	return &WfResult{
		Result: result,
	}, nil
}
```
you could write this

```go
func NiceWorkflow(ctx workflow.Context) (r *WfResult, err error) {
	var a ActivitiesStub
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	// a.DoSomething(ctx, workflow.GetInfo(ctx).WorkflowExecution.ID, "maybe a string?") 🪲 wrong argument!
	result, err := a.DoSomethingExec(ctx, workflow.GetInfo(ctx).WorkflowExecution.ID, &echo.Group{})
	if err != nil {
		return nil, err
	}
	return &WfResult{
		Result: result,
	}, nil
}
```

This tool accomplishes this by generating some simple stubs for you. You can have a look at [the test stub](./test/activities.gen.go) to see what is generated.

## Usage

Install with `go install github.com/Hades32/temporalgen@latest`

Now, just add a Go generate comment with the type you want to create stubs for
```go
//go:generate temporalgen -type Activities
type Activities struct {
	...
}
```

Now you can run `go generate` in your package (or `go generate ./...` for your whole module) and a `activities.gen.go` file will be generated.

I suggest running generate locally and checking the result into source control like any other source-code.
