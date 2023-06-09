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

	// merge metrics
	for _, spotMetric := range event.Metrics {
		added := false
		for _, aggregateMetric := range e.Metrics {
			if aggregateMetric.Key == spotMetric.Key {
				aggregateMetric.Value += spotMetric.Value
				added = true
				break
			}
		}
		if !added {
			e.Metrics = append(e.Metrics, spotMetric)
		}
	}

	// merge metadata
	for _, spotMeta := range event.Metadata {
		added := false
		for _, aggregateMeta := range e.Metadata {
			if aggregateMeta.Key == spotMeta.Key {
				aggregateMeta.Value = spotMeta.Value //replace
				added = true
				break
			}
		}
		if !added {
			e.Metadata = append(e.Metadata, spotMeta)
		}
	}

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
		event.Service +
		event.Network +
		event.ApiKeyUsage +
		event.ApiKeyId +
		event.IpAddress +
		event.Method
}
