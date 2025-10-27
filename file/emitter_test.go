package file

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/streamingfast/dmetering"
	"github.com/stretchr/testify/assert"
)

func newEvent(metricsKey string, metricsValue float64) dmetering.Event {
	return dmetering.Event{
		Endpoint: "sf.firehose.v1/Blocks",
		Metrics: map[string]float64{
			metricsKey: metricsValue,
		},
		OrganizationID: "0bizy1111111111111111",
		ApiKeyID:       "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress:      "192.168.1.1",
		Meta:           "test",
		Timestamp:      time.Now(),
	}
}

func newTestEmitter() (dmetering.EventEmitter, error) {
	return newEmitter("file:///tmp/test?buffer=5&flushIntervalSeconds=1&network=testNetwork", zap.NewNop(), newMockFS())
}

func TestEmitter(t *testing.T) {
	ctx := context.Background()

	e, _ := newTestEmitter()
	e.Emit(ctx, newEvent("egress_bytes", 100))
	e.Emit(ctx, newEvent("egress_bytes", 200))
	e.Emit(ctx, newEvent("egress_bytes", 300))

	time.Sleep(2 * time.Second)

	e.Emit(ctx, newEvent("egress_bytes", 100))
	e.Emit(ctx, newEvent("egress_bytes", 200))
	e.Emit(ctx, newEvent("egress_bytes", 300))

	time.Sleep(2 * time.Second)

	e.Shutdown(nil)
	fe, ok := e.(*emitter)
	if !ok {
		t.Fatal("Failed to cast to emitter")
	}
	files := fe.fs.(*mockFS).files
	assert.Equal(t, 2, len(files))
}
