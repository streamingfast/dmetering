package dmetering

import (
	"context"
)

var defaultMeter EventEmitter = newNullEmitter()

func SetDefaultEmitter(m EventEmitter) {
	defaultMeter = m
}

func Emit(ctx context.Context, event Event) error {
	return defaultMeter.Emit(ctx, event)
}
