package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/streamingfast/dmetering"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var zlog, tracer = logging.PackageLogger("dmetring", "github.com/streamingfast/dmetering/grpc.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zapcore.InfoLevel))
}

type mockClient struct {
	batchesInfo []int
	totalBytes  uint64
}

func (c *mockClient) Emit(ctx context.Context, in *pbmetering.Events, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.batchesInfo = append(c.batchesInfo, len(in.Events))
	for _, event := range in.Events {
		c.totalBytes += uint64(event.Metrics[0].Value)
	}
	return nil, nil
}

func (c *mockClient) Close() error {
	return nil
}

func newEvent(metricsKey string, metricsValue float64) dmetering.Event {
	return dmetering.Event{
		Endpoint: "sf.firehose.v1/Blocks",
		Metrics: map[string]float64{
			metricsKey: metricsValue,
		},
		UserID:    "0bizy1111111111111111",
		ApiKeyID:  "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}
}

func TestAuthenticatorPlugin_ContinuousAuthenticate(t *testing.T) {

	tests := []struct {
		name              string
		eventCount        int
		expectBatchesInfo []int
		expectTotalBytes  uint64
	}{
		{
			name:       "sending events",
			eventCount: 12,
			expectBatchesInfo: []int{
				5,
				5,
				2,
			},
			expectTotalBytes: 78,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			eventClient := &mockClient{}

			config := &Config{
				Endpoint:   "localhost:9000",
				BatchSize:  5,
				BufferSize: 100,
				Network:    "eth-testnet",
			}
			plugin, err := newWithClient(config, eventClient, eventClient.Close, zlog)
			require.NoError(t, err)

			for i := 0; i < test.eventCount; i++ {
				plugin.Emit(ctx, newEvent("read_bytes", float64(i+1)))
			}

			time.Sleep(1 * time.Second)
			plugin.Shutdown(nil)
			assert.Equal(t, test.expectBatchesInfo, eventClient.batchesInfo)
			assert.Equal(t, test.expectTotalBytes, eventClient.totalBytes)
		})
	}
}
