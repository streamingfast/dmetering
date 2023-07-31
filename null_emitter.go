package dmetering

import (
	"context"

	"go.uber.org/zap"
)

func RegisterNull() {
	Register("null", func(_ string, _ *zap.Logger) (EventEmitter, error) {
		return newNullEmitter(), nil
	})
}

type nullEmitter struct{}

func (e *nullEmitter) Emit(_ context.Context, _ Event) {}

func (e *nullEmitter) Shutdown(error) {}

func newNullEmitter() EventEmitter {
	return &nullEmitter{}
}
