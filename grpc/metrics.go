package grpc

import "github.com/streamingfast/dmetrics"

var MetricSet = dmetrics.NewSet()
var DroppedEventCounter = MetricSet.NewCounter("dropped_event", "Counter of drop metering events")
