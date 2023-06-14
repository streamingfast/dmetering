package dmetering

import (
	"context"
	"fmt"

	"github.com/streamingfast/dgrpc"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type grpcEmitter struct {
	*shutter.Shutter
	client pbmetering.MeteringClient

	logger *zap.Logger
}

func newGRPCEmitter(endpoint string, logger *zap.Logger) (EventEmitter, error) {
	client, cleanUp, err := newMeteringClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to create external gRPC client %w", err)
	}

	e := &grpcEmitter{
		Shutter: shutter.New(),
		client:  client,
		logger:  logger,
	}

	e.OnTerminated(func(_ error) {
		cleanUp()
	})

	return e, nil
}

func (g *grpcEmitter) Emit(ctx context.Context, ev Event) error {
	pbevent := ev.ToProtoMeteringEvent()

	if pbevent.Service == "" {
		g.logger.Warn("events must contain service, dropping event", zap.Object("event", ev))
		return nil
	}

	_, err := g.client.Emit(ctx, pbevent)
	if err != nil {
		return fmt.Errorf("unable to emit event: %w", err)
	}
	return nil
}

func newMeteringClient(endpoint string) (pbmetering.MeteringClient, func(), error) {
	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions, grpc.WithInsecure())
	conn, err := dgrpc.NewExternalClient(endpoint, dialOptions...)
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to create external gRPC client")
	}

	client := pbmetering.NewMeteringClient(conn)
	f := func() {
		conn.Close()
	}
	return client, f, nil
}

type loggerEmitter struct {
	logger *zap.Logger
}

func (l *loggerEmitter) Emit(_ context.Context, event Event) error {
	l.logger.Info("emit", zap.Object("event", event))
	return nil
}

func (l *loggerEmitter) Shutdown(_ error) {}

func newLoggerEmitter(logger *zap.Logger) EventEmitter {
	return &loggerEmitter{
		logger: logger,
	}
}

type nullEmitter struct{}

func (p *nullEmitter) Emit(_ context.Context, _ Event) error {
	return nil
}

func (l *nullEmitter) Shutdown(_ error) {}

func newNullEmitter() EventEmitter {
	return &nullEmitter{}
}
