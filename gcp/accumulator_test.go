package metering_gcp

import (
	"testing"
	"time"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAccumulatorDelay(t *testing.T) {
	delay := 50 * time.Millisecond
	done := make(chan bool)
	now := time.Now()

	emitter := func(event *pbmetering.Event) {
		zlog.Info("sending event", zap.Reflect("event", event))
		assert.True(t, time.Since(now) >= delay)
		close(done)
	}

	accumulator := newAccumulator(emitter, delay, zlog)
	accumulator.emit(&pbmetering.Event{UserId: "user.id.3"})
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Time exceeded")
	}

	zlog.Info("accumulator determinated")
}

func TestAccumulator(t *testing.T) {

	delay := 1 * time.Minute
	cases := []struct {
		name          string
		numberOfEvent int
	}{
		{
			name:          "1 event",
			numberOfEvent: 1,
		},
		{
			name:          "2 events",
			numberOfEvent: 2,
		},
		{
			name:          "100 events",
			numberOfEvent: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			done := make(chan bool)
			emitter := func(event *pbmetering.Event) {
				assert.Equal(t, int64(1*c.numberOfEvent), event.RequestsCount)
				assert.Equal(t, int64(10*c.numberOfEvent), event.ResponsesCount)
				assert.Equal(t, int64(100*c.numberOfEvent), event.RateLimitHitCount)
				assert.Equal(t, int64(1000*c.numberOfEvent), event.IngressBytes)
				assert.Equal(t, int64(10000*c.numberOfEvent), event.EgressBytes)
				assert.Equal(t, int64(100000*c.numberOfEvent), event.IdleTime)
				close(done)
			}
			accumulator := newAccumulator(emitter, delay, zlog)

			for i := 0; i < c.numberOfEvent; i++ {
				accumulator.emit(&pbmetering.Event{
					UserId:            "user.id.1",
					RequestsCount:     1,
					ResponsesCount:    10,
					RateLimitHitCount: 100,
					IngressBytes:      1000,
					EgressBytes:       10000,
					IdleTime:          100000,
				})
			}
			accumulator.emitAccumulatedEvents()
			select {
			case <-done:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Time exceeded")
			}

		})
	}
}

func TestAccumulatorDiffentEventKey(t *testing.T) {
	delay := 1 * time.Minute
	done := make(chan bool)
	events := map[string]*pbmetering.Event{}
	emitter := func(event *pbmetering.Event) {
		events[event.UserId] = event
		if len(events) < 2 {
			return
		}
		close(done)
	}
	accumulator := newAccumulator(emitter, delay, zlog)

	accumulator.emit(&pbmetering.Event{
		UserId: "user.id.1a",
	})
	accumulator.emit(&pbmetering.Event{
		UserId: "user.id.2a",
	})
	accumulator.emitAccumulatedEvents()

	assert.NotNil(t, events["user.id.1a"])
	assert.NotNil(t, events["user.id.2a"])

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Time exceeded")
	}

}
