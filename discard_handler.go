package medusa

import (
	"context"
	"log/slog"
)

// This is a temporary solution until
// https://go-review.googlesource.com/c/go/+/626486
// gets available

type discardHandler struct{}

func (dh discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (dh discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (dh discardHandler) WithAttrs(as []slog.Attr) slog.Handler     { return dh }
func (dh discardHandler) WithGroup(name string) slog.Handler        { return dh }
