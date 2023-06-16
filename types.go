package dmetering

import (
	"context"
	"fmt"

	"github.com/streamingfast/dgrpc"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"go.uber.org/zap"
)

type grpcEmitter struct {
	client pbmetering.MeteringClient

	network   string
	closeFunc CloseFunc
	logger    *zap.Logger
}

func newGRPCEmitter(network string, endpoint string, logger *zap.Logger) (EventEmitter, error) {
	client, closeFunc, err := newMeteringClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to create external gRPC client %w", err)
	}

	e := &grpcEmitter{
		client:    client,
		network:   network,
		closeFunc: closeFunc,
		logger:    logger,
	}

	return e, nil
}

func (g *grpcEmitter) Close() error {
	return g.closeFunc()
}

func (g *grpcEmitter) Emit(ctx context.Context, ev Event) {
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
	conn, err := dgrpc.NewInternalClient(endpoint)
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
