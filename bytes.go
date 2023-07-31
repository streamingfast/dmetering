package dmetering

import (
	"context"
	"fmt"
	"sync"
)

type bytesMeterKey string

const contextKey = bytesMeterKey("bytesMeter")

func GetBytesMeter(ctx context.Context) Meter {
	if bm, ok := ctx.Value(contextKey).(Meter); ok {
		return bm
	}
	return NoopBytesMeter
}

func WithBytesMeter(ctx context.Context) context.Context {
	bm := NewBytesMeter()
	return WithExistingBytesMeter(ctx, bm)
}

func WithExistingBytesMeter(ctx context.Context, bm Meter) context.Context {
	return context.WithValue(ctx, contextKey, bm)
}

type Meter interface {
	AddBytesWritten(n int)
	AddBytesRead(n int)

	BytesWritten() uint64
	BytesRead() uint64

	BytesWrittenDelta() uint64
	BytesReadDelta() uint64
}

type meter struct {
	bytesWritten uint64
	bytesRead    uint64

	bytesWrittenDelta uint64
	bytesReadDelta    uint64

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

func (b *meter) AddBytesRead(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.bytesReadDelta += uint64(n)
	b.bytesRead += uint64(n)
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

type noopMeter struct{}

func (_ *noopMeter) AddBytesWritten(n int)     { return }
func (_ *noopMeter) AddBytesRead(n int)        { return }
func (_ *noopMeter) BytesWritten() uint64      { return 0 }
func (_ *noopMeter) BytesRead() uint64         { return 0 }
func (_ *noopMeter) BytesWrittenDelta() uint64 { return 0 }
func (_ *noopMeter) BytesReadDelta() uint64    { return 0 }

var NoopBytesMeter Meter = &noopMeter{}
