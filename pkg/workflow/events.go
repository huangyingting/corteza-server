package workflow

import "context"

type (
	endEvent   struct{}
	errorEvent struct{ message string }
)

var (
	_ Finalizer = &endEvent{}
	_ Finalizer = &errorEvent{}
)

func EndEvent() *endEvent                                         { return &endEvent{} }
func (t *endEvent) Finalize(_ context.Context, s Variables) error { return nil }
func (endEvent) NodeRef() string                                  { return "(end)" }

func ErrorEvent(message string) *errorEvent                         { return &errorEvent{message} }
func (t *errorEvent) Finalize(_ context.Context, s Variables) error { return nil }
func (errorEvent) NodeRef() string                                  { return "(error)" }
