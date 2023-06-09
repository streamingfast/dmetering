package dmetering

import (
	"github.com/streamingfast/dauth/authenticator"

	pbmetering "github.com/streamingfast/dmetering/pb/sf/metering/v1"
)

//
//// Event represents a metering event that needs to be recorded
//type Event struct {
//	Service        string            `json:"service"`
//	Metadata       map[string]string `json:"metadata"`
//	Kind           string            `json:"kind"`
//	Method         string            `json:"method"`
//	RequestsCount  int64             `json:"requests_count,omitempty"`
//	ResponsesCount int64             `json:"responses_count,omitempty"`
//	IngressBytes   int64             `json:"ingress_bytes,omitempty"`
//	EgressBytes    int64             `json:"egress_bytes,omitempty"`
//}

type Eventable interface {
	ToProto(creds authenticator.Credentials, network string) *pbmetering.Event
}

var _ Eventable = (*FirehoseEvent)(nil)

type FirehoseEvent struct {
	Metadata       map[string]string `json:"metadata"`
	Method         string            `json:"method"`
	RequestsCount  int64             `json:"requests_count,omitempty"`
	ResponsesCount int64             `json:"responses_count,omitempty"`
	IngressBytes   int64             `json:"ingress_bytes,omitempty"`
	EgressBytes    int64             `json:"egress_bytes,omitempty"`
	BytesWritten   int64             `json:"bytes_written,omitempty"`
	BytesRead      int64             `json:"bytes_read,omitempty"`
}

func (ev *FirehoseEvent) ToProto(creds authenticator.Credentials, network string) *pbmetering.Event {
	var metadataFields []*pbmetering.MetadataField
	for k, v := range ev.Metadata {
		metadataFields = append(metadataFields, &pbmetering.MetadataField{
			Key:   k,
			Value: v,
		})
	}
	identification := creds.Identification()
	return &pbmetering.Event{
		Service:  "firehose",
		Method:   ev.Method,
		Network:  network,
		Metadata: metadataFields,
		Metrics: []*pbmetering.Metric{
			{Key: pbmetering.Metric_REQUESTS_COUNT, Value: ev.RequestsCount},
			{Key: pbmetering.Metric_RESPONSES_COUNT, Value: ev.ResponsesCount},
			{Key: pbmetering.Metric_INGRESS_BYTES, Value: ev.IngressBytes},
			{Key: pbmetering.Metric_EGRESS_BYTES, Value: ev.EgressBytes},
			{Key: pbmetering.Metric_READ_BYTES, Value: ev.BytesRead},
			{Key: pbmetering.Metric_WRITTEN_BYTES, Value: ev.BytesWritten},
		},
		UserId:      identification.UserId,
		ApiKeyId:    identification.ApiKeyId,
		ApiKey:      identification.ApiKey,
		ApiKeyUsage: identification.ApiKeyUsage,
		IpAddress:   identification.IpAddress,
	}
}

type HTTPEvent struct {
	Service        string `json:"service"`
	Method         string `json:"method"`
	RequestsCount  int64  `json:"requests_count,omitempty"`
	ResponsesCount int64  `json:"responses_count,omitempty"`
	IngressBytes   int64  `json:"ingress_bytes,omitempty"`
	EgressBytes    int64  `json:"egress_bytes,omitempty"`
}

func (ev *HTTPEvent) ToProto(creds authenticator.Credentials, network string) *pbmetering.Event {
	identification := creds.Identification()
	return &pbmetering.Event{
		Service: "firehose",
		Method:  ev.Method,
		Network: network,
		Metrics: []*pbmetering.Metric{
			{Key: pbmetering.Metric_REQUESTS_COUNT, Value: ev.RequestsCount},
			{Key: pbmetering.Metric_RESPONSES_COUNT, Value: ev.ResponsesCount},
			{Key: pbmetering.Metric_INGRESS_BYTES, Value: ev.IngressBytes},
			{Key: pbmetering.Metric_EGRESS_BYTES, Value: ev.EgressBytes},
		},
		UserId:      identification.UserId,
		ApiKeyId:    identification.ApiKeyId,
		ApiKey:      identification.ApiKey,
		ApiKeyUsage: identification.ApiKeyUsage,
		IpAddress:   identification.IpAddress,
	}
}
