package metering_gcp

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/streamingfast/dauth/authenticator"
	"github.com/streamingfast/dmetering"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var traceEnabled = os.Getenv("TRACE") == "true"

func init() {
	dmetering.Register("cloud-gcp", func(config string, logger *zap.Logger) (dmetering.Metering, error) {
		u, err := url.Parse(config)
		if err != nil {
			return nil, err
		}

		vals := u.Query()
		networkID := vals.Get("networkId")
		if networkID == "" {
			return nil, fmt.Errorf("missing networkId query param to metering config")
		}

		var emitterDelay = 10 * time.Second
		emitterDelayString := vals.Get("emitterDelay")
		if emitterDelayString != "" {
			if d, err := time.ParseDuration(emitterDelayString); err == nil {
				emitterDelay = d
			}
		}

		project := u.Host
		if project == "" {
			return nil, fmt.Errorf("project not specified (as hostname)")
		}

		topic := strings.TrimLeft(u.Path, "/")
		if topic == "" {
			return nil, fmt.Errorf("topic not specified (as path component)")
		}

		warnOnErrors := vals.Get("warnOnErrors") == "true"

		return newMetering(networkID, project, topic, warnOnErrors, emitterDelay, nil, nil, logger), nil
	})
}

type meteringPlugin struct {
	network string

	topic              *pubsub.Topic
	warnOnPubSubErrors bool

	messagesCount atomic.Uint64
	errorCount    atomic.Uint64

	accumulator *Accumulator
	logger      *zap.Logger
}

type topicProviderFunc func(pubsubProject string, topicName string) *pubsub.Topic
type topicEmitterFunc func(e *pbmetering.Event)

func newMetering(network, pubSubProject, pubSubTopic string, warnOnPubSubErrors bool, emitterDelay time.Duration, topicProvider topicProviderFunc, topicEmitter topicEmitterFunc, logger *zap.Logger) *meteringPlugin {
	m := &meteringPlugin{
		network:            network,
		warnOnPubSubErrors: warnOnPubSubErrors,
		logger:             logger,
	}

	if topicProvider == nil {
		m.topic = defaultTopicProvider(pubSubProject, pubSubTopic)
	} else {
		m.topic = topicProvider(pubSubProject, pubSubTopic)
	}

	if topicEmitter == nil {
		m.accumulator = newAccumulator(m.defaultTopicEmitter, emitterDelay, logger)
	} else {
		m.accumulator = newAccumulator(topicEmitter, emitterDelay, logger)
	}

	logger.Info("dbilling is ready to emit")
	return m
}

func (m *meteringPlugin) isStreamingFastDomain(ctx context.Context) bool {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, dom := range md.Get(":authority") {
			if strings.Contains(dom, "streamingfast.io") {
				return true
			}
		}
		if traceEnabled {
			m.logger.Debug("not streamingfast", zap.Strings("authority", md.Get(":authority")))
		}
	} else {
		if traceEnabled {
			m.logger.Debug("could not event get metadata...")
		}
	}
	return false
}

func (m *meteringPlugin) EmitWithContext(ev dmetering.Event, ctx context.Context) {
	credentials := authenticator.GetCredentials(ctx)
	if m.isStreamingFastDomain(ctx) {
		ev.Service = fmt.Sprintf("sf:%s", ev.Service) // will give something like `sf:firehose` or `sf:dgraphql` ...
	}
	m.EmitWithCredentials(ev, credentials)
}

func (m *meteringPlugin) EmitWithCredentials(ev dmetering.Event, creds authenticator.Credentials) {

	fields := []*pbmetering.MetadataField{}
	for k, v := range ev.Metadata {
		fields = append(fields, &pbmetering.MetadataField{
			Key:   k,
			Value: v,
		})
	}

	identification := creds.Identification()

	meteringeEvent := &pbmetering.Event{
		Service:           ev.Service,
		Kind:              ev.Kind,
		Method:            ev.Method,
		Network:           m.network,
		Metadata:          fields,
		RequestsCount:     ev.RequestsCount,
		ResponsesCount:    ev.ResponsesCount,
		RateLimitHitCount: ev.RateLimitHitCount,
		IngressBytes:      ev.IngressBytes,
		EgressBytes:       ev.EgressBytes,
		IdleTime:          ev.IdleTime,
		UserId:            identification.UserId,
		ApiKeyId:          identification.ApiKeyId,
		ApiKey:            identification.ApiKey,
		ApiKeyUsage:       identification.ApiKeyUsage,
		IpAddress:         identification.IpAddress,
	}

	m.emit(meteringeEvent)
}

func (m *meteringPlugin) emit(e *pbmetering.Event) {
	m.messagesCount.Inc()
	if e.Timestamp == nil {
		e.Timestamp = ptypes.TimestampNow()
	}
	m.accumulator.emit(e)
}

func (m *meteringPlugin) GetStatusCounters() (total, errors uint64) {
	return m.messagesCount.Load(), m.errorCount.Load()
}

func (m *meteringPlugin) WaitToFlush() {
	m.logger.Info("gracefully shutting down, now flushing pending dbilling events")
	m.accumulator.emitAccumulatedEvents()
	m.topic.Stop()
	m.logger.Info("all billing events have been flushed before shutdown")
}

func defaultTopicProvider(pubsubProject string, topicName string) *pubsub.Topic {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, pubsubProject)
	if err != nil {
		panic("unable to setup metering PubSub connection")
	}

	topic := client.Topic(topicName)
	topic.PublishSettings = pubsub.PublishSettings{
		ByteThreshold:  20000,
		CountThreshold: 100,
		DelayThreshold: 1 * time.Second,
	}

	exists, err := topic.Exists(ctx)
	if err != nil || !exists {

		panic(fmt.Errorf("unable to setup metering PubSub connection project: %s, topic: %s: %w", pubsubProject, topicName, err))
	}
	return topic
}

func (m *meteringPlugin) defaultTopicEmitter(e *pbmetering.Event) {
	if e.UserId == "" || e.Service == "" || e.Kind == "" {
		m.logger.Warn("events SHALL minimally contain UserID, Source and Kind, dropping billing event")
		return
	}

	data, err := proto.Marshal(e)
	if err != nil {
		m.errorCount.Inc()
		return
	}

	m.logger.Debug("sending message", zap.String("data_hex", hex.EncodeToString(data)))

	res := m.topic.Publish(context.Background(), &pubsub.Message{
		Data: data,
	})

	if m.warnOnPubSubErrors {
		_, err = res.Get(context.Background())
		if err != nil {
			m.logger.Warn("failed to publish", zap.Error(err))
		}
	}
}
