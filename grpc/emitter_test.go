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

var zlog, tracer = logging.PackageLogger("dmetering", "github.com/streamingfast/dmetering/grpc.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zapcore.DebugLevel))
}

type mockClient struct {
	eventCount     int
	totalBytes     uint64
	eventsReceived []*pbmetering.Event
}

func (c *mockClient) Emit(ctx context.Context, in *pbmetering.Events, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.eventCount += len(in.Events)
	for _, event := range in.Events {
		c.totalBytes += uint64(event.Metrics[0].Value)
	}
	if c.eventsReceived == nil {
		c.eventsReceived = make([]*pbmetering.Event, 0)
	}
	c.eventsReceived = append(c.eventsReceived, in.Events...)
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
		OrganizationID: "0bizy1111111111111111",
		ApiKeyID:       "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress:      "192.168.1.1",
		Meta:           "test",
		Timestamp:      time.Now(),
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
			name:             "sending events",
			eventCount:       12,
			expectTotalBytes: 78,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			eventClient := &mockClient{}

			config := &Config{
				Endpoint:   "localhost:9000",
				Delay:      100 * time.Millisecond,
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
			assert.Equal(t, test.eventCount, eventClient.eventCount)
			assert.Equal(t, test.expectTotalBytes, eventClient.totalBytes)
		})
	}
}

func TestNetworkOverride(t *testing.T) {

	ctx := context.Background()
	eventClient := &mockClient{}

	config := &Config{
		Endpoint:   "localhost:9000",
		Delay:      100 * time.Millisecond,
		BufferSize: 100,
		Network:    "eth-testnet",
	}
	plugin, err := newWithClient(config, eventClient, eventClient.Close, zlog)
	require.NoError(t, err)

	plugin.Emit(ctx, dmetering.Event{
		Endpoint: "sf.firehose.v1/Blocks",
		Metrics: map[string]float64{
			"requests": 1,
		},
		OrganizationID: "0bizy1111111111111111",
		ApiKeyID:       "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress:      "192.168.1.1",
		Network:        "eth-mainnet",
		Meta:           "test",
		Timestamp:      time.Now(),
	})

	plugin.Shutdown(nil)
	assert.Equal(t, 1, len(eventClient.eventsReceived))
	assert.Equal(t, "eth-testnet", eventClient.eventsReceived[0].Network)
}

func TestNetworkIncluded(t *testing.T) {

	ctx := context.Background()
	eventClient := &mockClient{}

	config := &Config{
		Endpoint:   "localhost:9000",
		Delay:      100 * time.Millisecond,
		BufferSize: 100,
	}
	plugin, err := newWithClient(config, eventClient, eventClient.Close, zlog)
	require.NoError(t, err)

	plugin.Emit(ctx, dmetering.Event{
		Endpoint: "sf.firehose.v1/Blocks",
		Metrics: map[string]float64{
			"requests": 1,
		},
		OrganizationID: "0bizy1111111111111111",
		ApiKeyID:       "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress:      "192.168.1.1",
		Network:        "eth-mainnet",
		Meta:           "test",
		Timestamp:      time.Now(),
	})

	plugin.Shutdown(nil)
	assert.Equal(t, 1, len(eventClient.eventsReceived))
	assert.Equal(t, "eth-mainnet", eventClient.eventsReceived[0].Network)
}

func TestNetworkMissing(t *testing.T) {

	ctx := context.Background()
	eventClient := &mockClient{}

	config := &Config{
		Endpoint:   "localhost:9000",
		Delay:      100 * time.Millisecond,
		BufferSize: 100,
	}
	plugin, err := newWithClient(config, eventClient, eventClient.Close, zlog)
	require.NoError(t, err)

	plugin.Emit(ctx, dmetering.Event{
		Endpoint: "sf.firehose.v1/Blocks",
		Metrics: map[string]float64{
			"requests": 1,
		},
		OrganizationID: "0bizy1111111111111111",
		ApiKeyID:       "2323232323232323232323232323232323232323232323232323232323232323",
		IpAddress:      "192.168.1.1",
		Meta:           "test",
		Timestamp:      time.Now(),
	})

	plugin.Shutdown(nil)
	assert.Equal(t, 0, len(eventClient.eventsReceived))
}
