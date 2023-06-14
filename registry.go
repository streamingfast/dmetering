package dmetering

import (
	"fmt"
	"net/url"

	"go.uber.org/zap"
)

var registry = make(map[string]FactoryFunc)

type FactoryFunc func(config string, logger *zap.Logger) (EventEmitter, error)

func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}

func RegisterGRPC() {
	Register("grpc", func(config string, logger *zap.Logger) (EventEmitter, error) {
		u, err := url.Parse(config)
		if err != nil {
			return nil, err
		}

		endpoint := u.Host
		if endpoint == "" {
			return nil, fmt.Errorf("endpoint not specified (as hostname)")
		}

		return newGRPCEmitter(endpoint, logger)
	})
}

func RegisterLogger() {
	Register("logger", func(_ string, logger *zap.Logger) (EventEmitter, error) {
		return newLoggerEmitter(logger), nil
	})
}

func RegisterNull() {
	Register("null", func(_ string, _ *zap.Logger) (EventEmitter, error) {
		return newNullEmitter(), nil
	})
}

func RegisterDefault() {
	RegisterGRPC()
	RegisterLogger()
	RegisterNull()
}
