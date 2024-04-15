package logger

import (
	"context"

	"github.com/streamingfast/dmetering"
	"go.uber.org/zap"
)

func Register() {
	dmetering.Register("logger", func(_ string, logger *zap.Logger) (dmetering.EventEmitter, error) {
		return newEmitter(logger), nil
	})
}

type emitter struct {
	logger *zap.Logger
}

func (l *emitter) Emit(_ context.Context, event dmetering.Event) {
	l.logger.Info("emit", zap.Object("event", event))
}

func (l *emitter) Shutdown(error) {}

func newEmitter(logger *zap.Logger) dmetering.EventEmitter {
	return &emitter{
		logger: logger,
	}
}
