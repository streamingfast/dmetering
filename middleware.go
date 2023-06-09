package dmetering

import (
	"net/http"
	"strings"
)

func NewMeteringMiddleware(next http.Handler, metering Metering, service string, trackRequestsAndResponses, trackIngressAndEgressBytes bool) http.Handler {
	return &MeteringMiddleware{
		next:                       next,
		metering:                   metering,
		service:                    service,
		trackRequestsAndResponses:  trackRequestsAndResponses,
		trackIngressAndEgressBytes: trackIngressAndEgressBytes,
	}
}

func NewMeteringMiddlewareFunc(metering Metering, service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewMeteringMiddleware(next, metering, service, true, true)
	}
}

func NewMeteringMiddlewareFuncWithOptions(metering Metering, service string, trackRequestsAndResponses, trackIngressAndEgressBytes bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewMeteringMiddleware(next, metering, service, trackRequestsAndResponses, trackIngressAndEgressBytes)
	}
}

type MeteringMiddleware struct {
	next                       http.Handler
	metering                   Metering
	service                    string
	kind                       string
	trackRequestsAndResponses  bool
	trackIngressAndEgressBytes bool
}

type MeteringResponseWriter struct {
	http.ResponseWriter
	totalWrittenBytes int64
}

func (m *MeteringMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We ignore WebSocket Upgrade requests (so we do not double count)
	if strings.ToLower(r.Header.Get("Connection")) != "upgrade" {
		meteringWriter := &MeteringResponseWriter{
			ResponseWriter: w,
		}

		m.next.ServeHTTP(meteringWriter, r)

		var req, resp uint64
		var in, out float64
		if m.trackRequestsAndResponses {
			req = 1
			resp = 1
		}
		if m.trackIngressAndEgressBytes {
			in = float64(r.ContentLength)
			out = float64(meteringWriter.totalWrittenBytes)
		}

		m.metering.EmitWithContext(
			Event{
				Service:        m.service,
				Method:         r.URL.Path,
				RequestsCount:  req,
				ResponsesCount: resp,
				IngressBytes:   in,
				EgressBytes:    out,
			}, r.Context())
	} else {
		m.next.ServeHTTP(w, r)
	}
}

func (w *MeteringResponseWriter) Write(data []byte) (int, error) {
	bytesOut, err := w.ResponseWriter.Write(data)
	w.totalWrittenBytes += int64(bytesOut)
	return bytesOut, err
}
