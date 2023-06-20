package dmetering

import (
	"context"
	"fmt"
	"time"

	"github.com/streamingfast/dgrpc"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type grpcEmitter struct {
	client pbmetering.MeteringClient

	eventBuffer   chan Event
	launchCtx     context.Context
	launchCancel  context.CancelFunc
	eventsDropped *atomic.Uint64

	network   string
	closeFunc CloseFunc
	logger    *zap.Logger
}

func newGRPCEmitter(network string, endpoint string, logger *zap.Logger, bufferSize int) (EventEmitter, error) {
	client, closeFunc, err := newMeteringClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to create external gRPC client %w", err)
	}

	launchCtx, launchCancel := context.WithCancel(context.Background())

	e := &grpcEmitter{
		client:    client,
		network:   network,
		closeFunc: closeFunc,
		logger:    logger,

		eventBuffer:   make(chan Event, bufferSize),
		launchCtx:     launchCtx,
		launchCancel:  launchCancel,
		eventsDropped: atomic.NewUint64(0),
	}

	go func() {
		for {
			select {
			case <-e.launchCtx.Done():
				return
			case ev := <-e.eventBuffer:
				e.emit(e.launchCtx, ev)
			}
		}
	}()

	go func() {
		//check every 10 seconds if there are dropped events
		for {
			select {
			case <-e.launchCtx.Done():
				return
			case <-time.After(10 * time.Second):
				dropped := e.eventsDropped.Load()
				if dropped > 0 {
					e.logger.Warn("metering events dropped. consider increasing buffer size", zap.Uint64("count", dropped), zap.Int("current_buffer_size", bufferSize))
				}
			}
		}
	}()

	return e, nil
}

func (g *grpcEmitter) Close() error {
	g.logger.Info("giving some time for buffer to clear")
	<-time.After(5 * time.Second)

	g.launchCancel()
	return g.closeFunc()
}

func (g *grpcEmitter) Emit(_ context.Context, ev Event) {
	select {
	case g.eventBuffer <- ev:
	default:
		g.eventsDropped.Inc()
	}
}

func (g *grpcEmitter) emit(ctx context.Context, ev Event) {
	pbevent := ev.ToProto(g.network)

	if pbevent.Endpoint == "" {
		g.logger.Warn("events must contain endpoint, dropping event", zap.Object("event", ev))
		return
	}

	_, err := g.client.Emit(ctx, pbevent)
	if err != nil {
		g.logger.Warn("failed to emit event", zap.Error(err))
	}
	return
}

func newMeteringClient(endpoint string) (pbmetering.MeteringClient, CloseFunc, error) {
	conn, err := dgrpc.NewInternalNoWaitClient(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create external gRPC client: %w", err)
	}

	client := pbmetering.NewMeteringClient(conn)
	return client, conn.Close, nil
}

type loggerEmitter struct {
	logger *zap.Logger
}

func (l *loggerEmitter) Emit(_ context.Context, event Event) {
	l.logger.Info("emit", zap.Object("event", event))
}

func (l *loggerEmitter) Close() error { return nil }

func newLoggerEmitter(logger *zap.Logger) EventEmitter {
	return &loggerEmitter{
		logger: logger,
	}
}

type nullEmitter struct{}

func (p *nullEmitter) Emit(_ context.Context, _ Event) {
}

func (l *nullEmitter) Close() error { return nil }

func newNullEmitter() EventEmitter {
	return &nullEmitter{}
}
