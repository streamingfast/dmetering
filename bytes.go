package dmetering

import (
	"context"
	"fmt"
	"sync"

	"github.com/streamingfast/logging"
	tracing "github.com/streamingfast/sf-tracing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type bytesMeterKey string

const contextKey = bytesMeterKey("bytesMeter")

func GetBytesMeter(ctx context.Context) Meter {
	if bm, ok := ctx.Value(contextKey).(Meter); ok && bm != nil {
		return bm
	}

	return NoopBytesMeter
}

func WithBytesMeter(ctx context.Context) context.Context {
	//check if meter already exists and that it is not nil or a noop
	if bm, ok := ctx.Value(contextKey).(Meter); ok && bm != nil && bm != NoopBytesMeter {
		return ctx
	}

	bm := NewBytesMeter()
	return WithExistingBytesMeter(ctx, bm)
}

func WithCounter(ctx context.Context, name string) context.Context {
	bm := GetBytesMeter(ctx)
	bm.AddCounter(name)
	return WithExistingBytesMeter(ctx, bm)
}

func WithExistingBytesMeter(ctx context.Context, bm Meter) context.Context {
	if bm == nil {
		return ctx
	}

	return context.WithValue(ctx, contextKey, bm)
}

type Meter interface {
	AddBytesWritten(n int)
	AddBytesRead(n int)

	AddBytesWrittenCtx(ctx context.Context, n int)
	AddBytesReadCtx(ctx context.Context, n int)

	BytesWritten() uint64
	BytesRead() uint64

	BytesWrittenDelta() uint64
	BytesReadDelta() uint64

	AddCounter(name string)
	CountInc(name string, n int)
	CountDec(name string, n int)
	GetCount(name string) int
	ResetCount(name string)
}

type meter struct {
	bytesWritten uint64
	bytesRead    uint64

	bytesWrittenDelta uint64
	bytesReadDelta    uint64

	counterMap map[string]int

	mu sync.RWMutex
}

func NewBytesMeter() Meter {
	return &meter{}
}

func (b *meter) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return fmt.Sprintf("bytes written: %d, bytes read: %d", b.bytesWritten, b.bytesRead)
}

func (b *meter) AddBytesWritten(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if n < 0 {
		panic("negative value")
	}

	b.bytesWrittenDelta += uint64(n)
	b.bytesWritten += uint64(n)
}

type mode string

const (
	modeRead  mode = "read"
	modeWrite mode = "write"
)

func (b *meter) AddBytesRead(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.bytesReadDelta += uint64(n)
	b.bytesRead += uint64(n)
}

func (b *meter) AddBytesWrittenCtx(ctx context.Context, n int) {
	logDataFromCtx(ctx, n, modeWrite)
	b.AddBytesWritten(n)
}

func (b *meter) AddBytesReadCtx(ctx context.Context, n int) {
	logDataFromCtx(ctx, n, modeRead)
	b.AddBytesRead(n)
}

func (b *meter) BytesWritten() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.bytesWritten
}

func (b *meter) BytesRead() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.bytesRead
}

func (b *meter) BytesWrittenDelta() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := b.bytesWrittenDelta
	b.bytesWrittenDelta = 0
	return result
}

func (b *meter) BytesReadDelta() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := b.bytesReadDelta
	b.bytesReadDelta = 0
	return result
}

func (b *meter) AddCounter(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.counterMap == nil {
		b.counterMap = make(map[string]int)
	}

	if _, ok := b.counterMap[name]; ok {
		return
	}

	b.counterMap[name] = 0
}

func (b *meter) CountInc(name string, n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.counterMap[name]; !ok {
		b.counterMap[name] = 0
	}

	b.counterMap[name] += n
}

func (b *meter) CountDec(name string, n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.counterMap[name]; !ok {
		b.counterMap[name] = 0
	}

	b.counterMap[name] -= n
}

func (b *meter) GetCount(name string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if val, ok := b.counterMap[name]; ok {
		return val
	}

	return 0
}

func (b *meter) ResetCount(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.counterMap[name]; ok {
		b.counterMap[name] = 0
	}

	return
}

type noopMeter struct{}

func (_ *noopMeter) ResetCount(name string)                        { return }
func (_ *noopMeter) AddBytesWritten(n int)                         { return }
func (_ *noopMeter) AddBytesRead(n int)                            { return }
func (_ *noopMeter) AddBytesWrittenCtx(ctx context.Context, n int) {}
func (_ *noopMeter) AddBytesReadCtx(ctx context.Context, n int)    {}
func (_ *noopMeter) BytesWritten() uint64                          { return 0 }
func (_ *noopMeter) BytesRead() uint64                             { return 0 }
func (_ *noopMeter) BytesWrittenDelta() uint64                     { return 0 }
func (_ *noopMeter) BytesReadDelta() uint64                        { return 0 }
func (_ *noopMeter) AddCounter(name string)                        { return }
func (_ *noopMeter) CountInc(name string, n int)                   { return }
func (_ *noopMeter) CountDec(name string, n int)                   { return }
func (_ *noopMeter) GetCount(name string) int                      { return 0 }

var NoopBytesMeter Meter = &noopMeter{}

type MeterLogData struct {
	Filename *string
	Store    *string
	TraceId  *string
	Mode     mode
	NBytes   int
}

func (mld *MeterLogData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if mld.Filename != nil {
		enc.AddString("file", *mld.Filename)
	}
	if mld.Store != nil {
		enc.AddString("store", *mld.Store)
	}
	if mld.TraceId != nil {
		enc.AddString("trace_id", *mld.TraceId)
	}
	enc.AddString("mode", string(mld.Mode))
	enc.AddInt("bytes", mld.NBytes)

	return nil
}

func logDataFromCtx(ctx context.Context, nBytes int, mode mode) {
	var logger *zap.Logger
	if val, ok := ctx.Value("logger").(*zap.Logger); ok {
		logger = val
	} else {
		return
	}

	var tracer logging.Tracer
	if val, ok := ctx.Value("tracer").(logging.Tracer); ok {
		tracer = val
	} else {
		return
	}

	var filename, store, traceId *string
	if val, ok := ctx.Value("file").(string); ok {
		filename = &val
	}
	if val, ok := ctx.Value("store").(string); ok {
		store = &val
	}

	traceIdVal := tracing.GetTraceID(ctx).String()
	traceId = &traceIdVal

	mld := &MeterLogData{
		Filename: filename,
		Store:    store,
		TraceId:  traceId,
		Mode:     mode,
		NBytes:   nBytes,
	}

	if tracer.Enabled() {
		logger.Debug("bytes metering", zap.Object("meterLogData", mld))
	}
}
