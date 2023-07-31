package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/streamingfast/shutter"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/dmetering"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"go.uber.org/zap"
)

func Register() {
	dmetering.Register("grpc", func(config string, logger *zap.Logger) (dmetering.EventEmitter, error) {
		c, err := newConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config string %s: %w", config, err)
		}
		return new(c, logger)
	})
}

type CloseFunc func() error

type emitter struct {
	*shutter.Shutter
	config *Config

	activeBatch     []*pbmetering.Event
	buffer          chan dmetering.Event
	client          pbmetering.MeteringClient
	clientCloseFunc CloseFunc
	done            chan bool

	logger *zap.Logger
}

func new(config *Config, logger *zap.Logger) (dmetering.EventEmitter, error) {
	client, closeFunc, err := newMeteringClient(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to create external gRPC client %w", err)
	}

	return newWithClient(config, client, closeFunc, logger)
}

func newWithClient(
	config *Config,
	client pbmetering.MeteringClient,
	closeFunc CloseFunc,
	logger *zap.Logger,
) (dmetering.EventEmitter, error) {
	e := &emitter{
		Shutter:         shutter.New(),
		config:          config,
		client:          client,
		clientCloseFunc: closeFunc,
		buffer:          make(chan dmetering.Event, config.BufferSize),
		activeBatch:     []*pbmetering.Event{},
		done:            make(chan bool, 1),
		logger:          logger.Named("metrics.emitter"),
	}

	go e.launch()

	e.OnTerminating(func(err error) {
		e.logger.Info("received shutdown signal, waiting for launch loop to end", zap.Error(err))
		<-e.done
		e.flushAndCloseEvent()
		e.clientCloseFunc()

	})
	return e, nil
}

func (e *emitter) launch() {
	for {
		select {
		case <-e.Terminating():
			e.done <- true
		case ev := <-e.buffer:
			protoEv := ev.ToProto(e.config.Network)
			if len(e.activeBatch) == int(e.config.BatchSize) {
				e.emit(e.activeBatch)
				e.activeBatch = []*pbmetering.Event{protoEv}
			} else {
				e.activeBatch = append(e.activeBatch, protoEv)
			}
		}
	}
}

func (e *emitter) flushAndCloseEvent() {
	close(e.buffer)

	t0 := time.Now()
	e.logger.Info("waiting for event flush to complete", zap.Int("count", len(e.buffer)))
	defer func() {
		e.logger.Info("event flushed", zap.Duration("elapsed", time.Since(t0)))
	}()

	for {
		ev, ok := <-e.buffer
		protoEv := ev.ToProto(e.config.Network)
		if !ok {
			e.logger.Info("sending last events", zap.Int("count", len(e.activeBatch)))
			e.emit(e.activeBatch)
			return
		}
		e.activeBatch = append(e.activeBatch, protoEv)
	}
}

func (e *emitter) Emit(_ context.Context, ev dmetering.Event) {
	if ev.Endpoint == "" {
		e.logger.Warn("events must contain endpoint, dropping event", zap.Object("event", ev))
		return
	}

	if e.IsTerminating() {
		e.logger.Warn("emitter is shutting down cannot track event", zap.Object("event", ev))
		return
	}

	select {
	case e.buffer <- ev:
	default:
		if e.config.PanicOnDrop {
			panic(fmt.Errorf("failed to queue metric channel is full"))
		}
		DroppedEventCounter.Inc()
	}
}

func (e *emitter) emit(events []*pbmetering.Event) {
	e.logger.Debug("tracking events", zap.Int("count", len(events)))
	if _, err := e.client.Emit(context.Background(), &pbmetering.Events{Events: events}); err != nil {
		e.logger.Warn("failed to emit event", zap.Error(err))
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
