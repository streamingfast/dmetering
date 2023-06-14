package dmetering

import (
	"context"
	"fmt"
	"net/url"
	"time"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Event struct {
	Network string             `json:"network"`
	Service string             `json:"service"`
	Metrics map[string]float64 `json:"metrics,omitempty"`

	UserID    string `json:"user_id"`
	ApiKeyID  string `json:"api_key_id"`
	IpAddress string `json:"ip_address"`

	Timestamp time.Time `json:"timestamp"`
}

func (ev Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("user_id", ev.UserID)
	enc.AddString("api_key_id", ev.ApiKeyID)
	enc.AddString("ip_address", ev.IpAddress)
	enc.AddString("endpoint", ev.Service)
	enc.AddString("network", ev.Network)
	enc.AddTime("timestamp", ev.Timestamp)

	for k, v := range ev.Metrics {
		enc.AddFloat64(k, v)
	}

	return nil
}

func (ev Event) ToProtoMeteringEvent() *pbmetering.Event {
	pbev := new(pbmetering.Event)
	pbev.Service = ev.Service
	pbev.Network = ev.Network
	pbev.Timestamp = timestamppb.New(ev.Timestamp)
	pbev.UserId = ev.UserID
	pbev.ApiKeyId = ev.ApiKeyID
	pbev.IpAddress = ev.IpAddress

	pbev.Metrics = []*pbmetering.Metric{}
	for k, v := range ev.Metrics {
		pbev.Metrics = append(pbev.Metrics, &pbmetering.Metric{
			Key:   k,
			Value: v,
		})
	}

	return pbev
}

type EventEmitter interface {
	Shutdown(error)
	Emit(ctx context.Context, ev Event) error
}

func New(config string, logger *zap.Logger) (EventEmitter, error) {
	u, err := url.Parse(config)
	if err != nil {
		return nil, err
	}

	factory := registry[u.Scheme]
	if factory == nil {
		panic(fmt.Sprintf("no Metering plugin named \"%s\" is currently registered.", u.Scheme))
	}
	return factory(config, logger)
}
