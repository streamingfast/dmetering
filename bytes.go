package dmetering

import (
	"context"
	"fmt"
	"sync"
)

type bytesMeterKey string

const contextKey = bytesMeterKey("bytesMeter")

func GetBytesMeter(ctx context.Context) BytesMeter {
	if bm, ok := ctx.Value(contextKey).(BytesMeter); ok {
		return bm
	}
	return NoopBytesMeter
}

func WithBytesMeter(ctx context.Context, bm BytesMeter) context.Context {
	return context.WithValue(ctx, contextKey, bm)
}

type BytesMeter interface {
	AddBytesWritten(n int)
	AddBytesRead(n int)

	BytesWritten() uint64
	BytesRead() uint64

	BytesWrittenDelta() uint64
	BytesReadDelta() uint64
}

type bytesMeter struct {
	bytesWritten uint64
	bytesRead    uint64

	bytesWrittenDelta uint64
	bytesReadDelta    uint64

	mu sync.RWMutex
}

func NewBytesMeter() BytesMeter {
	return &bytesMeter{}
}

func (b *bytesMeter) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return fmt.Sprintf("bytes written: %d, bytes read: %d", b.bytesWritten, b.bytesRead)
}

func (b *bytesMeter) AddBytesWritten(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if n < 0 {
		panic("negative value")
	}

	b.bytesWrittenDelta += uint64(n)
	b.bytesWritten += uint64(n)
}

func (b *bytesMeter) AddBytesRead(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.bytesReadDelta += uint64(n)
	b.bytesRead += uint64(n)
}

func (b *bytesMeter) BytesWritten() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.bytesWritten
}

func (b *bytesMeter) BytesRead() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.bytesRead
}

func (b *bytesMeter) BytesWrittenDelta() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := b.bytesWrittenDelta
	b.bytesReadDelta = 0
	return result
}

func (b *bytesMeter) BytesReadDelta() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := b.bytesReadDelta
	b.bytesReadDelta = 0
	return result
}

type noopBytesMeter struct{}

func (_ *noopBytesMeter) AddBytesWritten(n int)     { return }
func (_ *noopBytesMeter) AddBytesRead(n int)        { return }
func (_ *noopBytesMeter) BytesWritten() uint64      { return 0 }
func (_ *noopBytesMeter) BytesRead() uint64         { return 0 }
func (_ *noopBytesMeter) BytesWrittenDelta() uint64 { return 0 }
func (_ *noopBytesMeter) BytesReadDelta() uint64    { return 0 }

var NoopBytesMeter BytesMeter = &noopBytesMeter{}
