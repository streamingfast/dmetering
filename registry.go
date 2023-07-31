package dmetering

import (
	"go.uber.org/zap"
)

var registry = make(map[string]FactoryFunc)

type FactoryFunc func(config string, logger *zap.Logger) (EventEmitter, error)

func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}
