package metering_gcp

import (
	"sync"
	"time"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
)

type Accumulator struct {
	events       map[string]*pbmetering.Event
	eventsLock   sync.Mutex
	emitter      topicEmitterFunc
	emitterDelay time.Duration
	logger       *zap.Logger
}

func newAccumulator(emitter func(event *pbmetering.Event), emitterDelay time.Duration, logger *zap.Logger) *Accumulator {
	accumulator := &Accumulator{
		events:       make(map[string]*pbmetering.Event),
		emitter:      emitter,
		emitterDelay: emitterDelay,
		logger:       logger,
	}
	go accumulator.delayedEmitter()
	return accumulator
}

func (a *Accumulator) emit(event *pbmetering.Event) {
	if traceEnabled {
		a.logger.Debug("accumulator emitting", zap.String("user_id", event.UserId), zap.String("service", event.Service))
	}
	a.eventsLock.Lock()
	defer a.eventsLock.Unlock()

	key := eventToKey(event)
	e := a.events[key]

	if e == nil {
		a.events[key] = event
		return
	}

	e.RequestsCount += event.RequestsCount
	e.ResponsesCount += event.ResponsesCount
	e.RateLimitHitCount += event.RateLimitHitCount
	e.IngressBytes += event.IngressBytes
	e.EgressBytes += event.EgressBytes
	e.IdleTime += event.IdleTime
	e.Timestamp = ptypes.TimestampNow()

}

func (a *Accumulator) delayedEmitter() {
	for {
		time.Sleep(a.emitterDelay)
		a.logger.Debug("accumulator sleep over")
		a.emitAccumulatedEvents()
	}
}

func (a *Accumulator) emitAccumulatedEvents() {
	a.logger.Debug("emitting accumulated events")
	a.eventsLock.Lock()
	toSend := a.events
	a.events = make(map[string]*pbmetering.Event)
	a.eventsLock.Unlock()

	for _, event := range toSend {
		a.emitter(event)
	}
}

func eventToKey(event *pbmetering.Event) string {
	//UserId               string
	//Kind                 string
	//Source               string
	//Network              string
	//Usage                string
	//ApiKeyId             string
	//IpAddress            string -- optional -- collected circa 2020-01-20
	//Method               string -- optional -- collected circa 2020-02-18

	// Accumulator `GROUP BY` key
	return event.UserId +
		event.Kind +
		event.Service +
		event.Network +
		event.ApiKeyUsage +
		event.ApiKeyId +
		event.IpAddress +
		event.Method
}
