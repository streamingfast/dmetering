package dmetering

// Event represents a metering event that needs to be recorded
type Event struct {
	Source            string `json:"source"`
	Kind              string `json:"kind"`
	Method            string `json:"method"`
	RequestsCount     int64  `json:"requests_count,omitempty"`
	ResponsesCount    int64  `json:"responses_count,omitempty"`
	IngressBytes      int64  `json:"ingress_bytes,omitempty"`
	EgressBytes       int64  `json:"egress_bytes,omitempty"`
	IdleTime          int64  `json:"idle_time,omitempty"`
	RateLimitHitCount int64  `json:"rate_limit_hit_count,omitempty"`
}
