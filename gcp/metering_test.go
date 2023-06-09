package metering_gcp

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"

	"cloud.google.com/go/pubsub"
	"github.com/streamingfast/dauth/authenticator"
	"github.com/streamingfast/dmetering"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var _ authenticator.Credentials = (*mockCredentials)(nil)

type mockCredentials struct {
	i *authenticator.Identification
}

func (m mockCredentials) GetLogFields() []zap.Field                     { return nil }
func (m mockCredentials) Identification() *authenticator.Identification { return m.i }
func (m mockCredentials) Features() *authenticator.Features             { return &authenticator.Features{} }

func TestEmitWithContext(t *testing.T) {
	done := make(chan bool)

	topicProvider := func(pubsubProject string, t string) *pubsub.Topic {
		return nil
	}
	topicEmitter := func(event *pbmetering.Event) {
		assert.Equal(t, "subject.1", event.UserId)
		assert.Equal(t, "api.key.1", event.ApiKeyId)
		assert.Equal(t, "usage.1", event.ApiKeyUsage)
		close(done)
	}

	m := newMetering("network", "proj", "topic", false, 10*time.Millisecond, topicProvider, topicEmitter, zap.NewNop())

	ctx := context.Background()
	ctx = authenticator.WithCredentials(ctx, &mockCredentials{
		i: &authenticator.Identification{
			UserId:      "subject.1",
			ApiKeyId:    "api.key.1",
			ApiKey:      "",
			ApiKeyUsage: "usage.1",
			IpAddress:   "",
		},
	})
	m.EmitWithContext(&dmetering.FirehoseEvent{}, ctx)
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Time exceeded")
	}

}

func TestEmitWithContextMissingCredentials(t *testing.T) {
	done := make(chan bool)

	topicProvider := func(pubsubProject string, t string) *pubsub.Topic {
		return nil
	}
	topicEmitter := func(event *pbmetering.Event) {
		assert.Equal(t, "anonymous", event.UserId)
		assert.Equal(t, "anonymous", event.ApiKeyId)
		assert.Equal(t, "anonymous", event.ApiKey)
		close(done)
	}

	m := newMetering("network.1", "P", "dev-billable-events-v2", false, 10*time.Millisecond, topicProvider, topicEmitter, zap.NewNop())
	m.EmitWithContext(&dmetering.FirehoseEvent{}, context.Background())

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Time exceeded")
	}

}

func TestEmitter(t *testing.T) {
	ctx := context.Background()
	c, _ := NewFakePubsub(ctx)
	topic, err := c.CreateTopic(context.Background(), "dev-billable-events-v2")
	require.NoError(t, err)

	sub, err := c.CreateSubscription(context.Background(), "dev-billable-events-v2", pubsub.SubscriptionConfig{Topic: topic})
	require.NoError(t, err)

	topicProvider := func(pubsubProject string, t string) *pubsub.Topic {
		return topic
	}

	m := newMetering("network.1", "P", "dev-billable-events-v2", false, 10*time.Millisecond, topicProvider, nil, zap.NewNop())

	done := make(chan bool)

	// Fake "bean-counter" subscriber
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
			zlog.Debug("received message", zap.String("data_hex", hex.EncodeToString(message.Data)))
			event := &pbmetering.Event{}
			require.NoError(t, proto.Unmarshal(message.Data, event))

			assert.Equal(t, "user.id.1", event.UserId)
			assert.Equal(t, "network.1", event.Network)
			assert.NotNil(t, event.Timestamp)

			mCount, eCount := m.GetStatusCounters()
			assert.Equal(t, uint64(1), mCount)
			assert.Equal(t, uint64(0), eCount)
			close(done)
		})
		require.NoError(t, err)
	}()

	m.EmitWithCredentials(&dmetering.FirehoseEvent{
		Method: "method.1",
	}, &mockCredentials{
		i: &authenticator.Identification{
			UserId: "user.id.1",
		},
	})

	m.WaitToFlush()
	zlog.Info("emitted, waiting")

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Time exceeded")
	}
}
