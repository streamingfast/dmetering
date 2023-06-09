package dmetering

// Event represents a metering event that needs to be recorded
type Event struct {
	// Defines the service that emitted the event (firehose, substreams, evmx ...)
	Service string `json:"service"`
	// Defines the method within the service  (grpc_blocks, grpc_run, http_get_state, ....)
	Method         string            `json:"method"`
	Metadata       map[string]string `json:"metadata"`
	RequestsCount  uint64            `json:"requests_count,omitempty"`
	ResponsesCount uint64            `json:"responses_count,omitempty"`
	IngressBytes   float64           `json:"ingress_bytes,omitempty"`
	EgressBytes    float64           `json:"egress_bytes,omitempty"`
	WrittenBytes   float64           `json:"written_bytes,omitempty"`
	ReadBytes      float64           `json:"read_bytes,omitempty"`
}
