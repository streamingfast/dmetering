package dmetering

import (
	"context"
	"github.com/dfuse-io/dauth"
)

var defaultMeter Metering = newNullPlugin()

func SetDefaultMeter(m Metering) {
	defaultMeter = m
}

func EmitWithContext(ev Event, ctx context.Context) {
	defaultMeter.EmitWithContext(ev, ctx)
}

func EmitWithCredentials(ev Event, creds dauth.Credentials) {
	defaultMeter.EmitWithCredentials(ev, creds)
}

func GetStatusCounters() (total, errors uint64) {
	return defaultMeter.GetStatusCounters()
}

func WaitToFlush() {
	defaultMeter.WaitToFlush()
}
