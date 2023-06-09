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
			numberOfEvent: 100,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			done := make(chan bool)
			emitter := func(event *pbmetering.Event) {
				//check metrics
				for _, metric := range event.Metrics {
					var base int
					switch metric.Key {
					case "requests_count":
						base = 1
					case "responses_count":
						base = 10
					case "ingress_bytes":
						base = 100
					case "egress_bytes":
						base = 1000
					case "read_bytes":
						base = 10000
					case "written_bytes":
						base = 100000
					default:
						panic("unknown value")
					}
					assert.Equal(t, float64(base*c.numberOfEvent), metric.Value)
				}

				//check metadata
				assert.Equal(t, 2, len(event.Metadata))

				close(done)
			}
			accumulator := newAccumulator(emitter, delay, zlog)

			for i := 0; i < c.numberOfEvent; i++ {
				accumulator.emit(&pbmetering.Event{
					UserId: "user.id.1",
					Metrics: []*pbmetering.Metric{
						{Key: "requests_count", Value: 1},
						{Key: "responses_count", Value: 10},
						{Key: "ingress_bytes", Value: 100},
						{Key: "egress_bytes", Value: 1000},
						{Key: "read_bytes", Value: 10000},
						{Key: "written_bytes", Value: 100000},
					},
					Metadata: []*pbmetering.MetadataField{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
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

func TestAccumulatorDifferentEventKey(t *testing.T) {
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
