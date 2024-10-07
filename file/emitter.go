package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/streamingfast/dmetering"
	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
)

func Register() {
	dmetering.Register("file", func(configURL string, logger *zap.Logger) (dmetering.EventEmitter, error) {
		return newEmitter(configURL, logger, OSFS{})
	})
}

type emitter struct {
	*shutter.Shutter

	config *Config

	activeBatch []*pbmetering.Event
	buffer      chan dmetering.Event

	logger *zap.Logger
	done   chan bool
	lock   sync.Mutex

	fs FS
}

func newEmitter(configURL string, logger *zap.Logger, fs FS) (dmetering.EventEmitter, error) {
	config, err := newConfig(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config string %s: %w", configURL, err)
	}

	fe := &emitter{
		Shutter: shutter.New(),

		config: config,

		done:        make(chan bool),
		buffer:      make(chan dmetering.Event, config.BufferSize),
		activeBatch: []*pbmetering.Event{},
		logger:      logger.Named("file.metrics.emitter"),

		fs: fs,
	}

	fe.OnTerminating(func(_ error) {
		<-fe.done
		err := fe.flush()
		if err != nil {
			logger.Error("failed to flush on shutdown", zap.Error(err))
		}
	})

	go fe.launch()

	return fe, nil
}

func (e *emitter) launch() {
	ticker := time.NewTicker(time.Duration(e.config.FlushIntervalSeconds) * time.Second)
	for {
		select {
		case <-e.Terminating():
			e.done <- true
			return
		case <-ticker.C:
			err := e.flush()
			if err != nil {
				e.logger.Error("failed to flush", zap.Error(err))
			}
		case ev := <-e.buffer:
			e.activeBatch = append(e.activeBatch, ev.ToProto(e.config.Network))
		}
	}
}

func (e *emitter) Emit(_ context.Context, ev dmetering.Event) {
	if e.IsTerminating() {
		e.logger.Warn("emitter is shutting down cannot track event", zap.Object("event", ev))
		return
	}

	select {
	case e.buffer <- ev:
	default:
		e.logger.Warn("buffer is full, dropping event", zap.Object("event", ev))
	}
}

func (e *emitter) flush() error {
	if len(e.activeBatch) == 0 {
		return nil
	}

	var bs [][]byte
	for _, event := range e.activeBatch {
		b, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		bs = append(bs, b)
	}

	out := bytes.Join(bs, []byte("\n"))

	filename := e.fileName()

	f, err := e.fs.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	e.activeBatch = []*pbmetering.Event{}
	return nil
}

func (e *emitter) fileName() string {
	source := e.config.Source
	if e.config.Source == "" {
		e.config.Source = "unknown"
	}

	return filepath.Join(e.config.BasePath, fmt.Sprintf("events_%s_%d.jsonl", source, time.Now().Unix()))
}
