package dmetering

// Event represents a metering event that needs to be recorded
type Event struct {
	Source            string
	Kind              string
	Method            string
	RequestsCount     int64
	ResponsesCount    int64
	IngressBytes      int64
	EgressBytes       int64
	IdleTime          int64
	RateLimitHitCount int64
}
