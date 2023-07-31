package grpc

import "github.com/streamingfast/dmetrics"

var MetricSet = dmetrics.NewSet()
var DroppedEventCounter = MetricSet.NewCounter("dropped_event_counter", "Counter of drop metering events")
var MeteringGRPCErrCounter = MetricSet.NewCounter("metering_grpc_err_counter", "Counter of GRPC errors received")
