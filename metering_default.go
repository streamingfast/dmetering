package dmetering

import (
	"context"
)

var defaultMeter EventEmitter = newNullEmitter()

func SetDefaultEmitter(m EventEmitter) {
	defaultMeter = m
}

func GetDefaultEmitter() EventEmitter {
	return defaultMeter
}

func Emit(ctx context.Context, event Event) {
	defaultMeter.Emit(ctx, event)
}
