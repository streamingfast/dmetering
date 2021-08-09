package dmetering

import (
	"context"
	"fmt"
	"net/url"

	"github.com/streamingfast/dauth/authenticator"
)

type Metering interface {
	EmitWithContext(ev Event, ctx context.Context)
	EmitWithCredentials(ev Event, creds authenticator.Credentials)
	GetStatusCounters() (total, errors uint64)
	WaitToFlush()
}

var registry = make(map[string]FactoryFunc)

func New(config string) (Metering, error) {
	u, err := url.Parse(config)
	if err != nil {
		return nil, err
	}

	factory := registry[u.Scheme]
	if factory == nil {
		panic(fmt.Sprintf("no Metering plugin named \"%s\" is currently registered", u.Scheme))
	}
	return factory(config)
}

type FactoryFunc func(config string) (Metering, error)

func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}
