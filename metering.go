package dmetering

import (
	"context"
	"fmt"
	"net/url"

	"go.uber.org/zap"

	"github.com/streamingfast/dauth/authenticator"
)

type Metering interface {
	EmitWithContext(ev Eventable, ctx context.Context)
	EmitWithCredentials(ev Eventable, creds authenticator.Credentials)
	GetStatusCounters() (total, errors uint64)
	WaitToFlush()
}

var registry = make(map[string]FactoryFunc)

func New(config string, logger *zap.Logger) (Metering, error) {
	u, err := url.Parse(config)
	if err != nil {
		return nil, err
	}

	factory := registry[u.Scheme]
	if factory == nil {
		panic(fmt.Sprintf("no Metering plugin named \"%s\" is currently registered", u.Scheme))
	}
	return factory(config, logger)
}

type FactoryFunc func(config string, logger *zap.Logger) (Metering, error)

func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}
