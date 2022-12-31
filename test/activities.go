package test

import (
	"context"
	"encoding/json"
)

//go:generate temporalgen -type Activities
type Activities struct {
}

func (a *Activities) MarkReadyForUploads(_ context.Context, jobID string) (err error) {
	return nil
}

func (a *Activities) DoSomething(_ context.Context, jobID string, i *json.RawMessage) (string, error) {
	return "", nil
}
