# Stub Generation for Temporal Activities

## DEPRECATED

Since the introduction fo generics to Go, this repo is for less useful and as running it is releatively slow I now don't recommend it anymore.

Instead I started using simple generic helper methods that you can copy below. The only issue is that you must use a single struct as activity arguments, but this is actually the best way to ensure backwards compatibility anyway.

```go
func execActivityIO[Tin any, Tout any](ctx workflow.Context, activity func(ctx context.Context, params Tin) (res Tout, err error), input Tin, options ...execActivityOption) (Tout, error) {
	opts := workflow.GetActivityOptions(ctx)
	for _, opt := range options {
		opt(&opts)
	}
	ctx = workflow.WithActivityOptions(ctx, opts)
	f := workflow.ExecuteActivity(ctx, activity, input)
	var res Tout
	err := f.Get(ctx, &res)
	if err != nil {
		return res, fmt.Errorf("error in activity '%s': %w", utils.FuncName(activity), err)
	}
	return res, nil
}

func execActivity[Tin any](ctx workflow.Context, activity func(ctx context.Context, params Tin) (err error), input Tin, options ...execActivityOption) error {
	opts := workflow.GetActivityOptions(ctx)
	for _, opt := range options {
		opt(&opts)
	}
	ctx = workflow.WithActivityOptions(ctx, opts)
	f := workflow.ExecuteActivity(ctx, activity, input)
	err := f.Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("error in activity '%s': %w", utils.FuncName(activity), err)
	}
	return nil
}

type execActivityOption func(*workflow.ActivityOptions)

func startToClose(dur time.Duration) execActivityOption {
	return func(ao *workflow.ActivityOptions) {
		ao.StartToCloseTimeout = dur
		// fix consistency for overwrites
		if ao.ScheduleToCloseTimeout < dur {
			ao.ScheduleToCloseTimeout = dur
		}
	}
}

func scheduleToClose(dur time.Duration) execActivityOption {
	return func(ao *workflow.ActivityOptions) {
		ao.ScheduleToCloseTimeout = dur
	}
}
```

## OLD README BELOW

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
