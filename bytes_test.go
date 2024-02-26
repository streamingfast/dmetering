package dmetering

import (
	"context"
	"testing"
)

func TestMeter_BytesWrittenDelta(t *testing.T) {
	meter := NewBytesMeter()
	meter.AddBytesWritten(10)
	meter.AddBytesWritten(10)

	if meter.BytesWrittenDelta() != 20 {
		t.Errorf("expected 20, got %d", meter.BytesWrittenDelta())
	}

	meter.AddBytesWritten(10)

	if meter.BytesWrittenDelta() != 10 {
		t.Errorf("expected 10, got %d", meter.BytesWrittenDelta())
	}
}

func TestMeter_BytesReadDelta(t *testing.T) {
	meter := NewBytesMeter()
	meter.AddBytesRead(10)
	meter.AddBytesRead(10)

	if meter.BytesReadDelta() != 20 {
		t.Errorf("expected 20, got %d", meter.BytesReadDelta())
	}

	meter.AddBytesRead(10)

	if meter.BytesReadDelta() != 10 {
		t.Errorf("expected 10, got %d", meter.BytesReadDelta())
	}
}

func TestWithCounter(t *testing.T) {
	meter := NewBytesMeter()
	ctx := WithExistingBytesMeter(context.Background(), meter)
	ctx = WithCounter(ctx, "test")

	bm := GetBytesMeter(ctx)
	bm.CountInc("test", 10)

	c := meter.GetCount("test")
	if c != 10 {
		t.Errorf("expected 10, got %d", c)
	}
}

func TestWithBytesMeter(t *testing.T) {
	ctx := WithBytesMeter(context.Background())
	if GetBytesMeter(ctx) == nil {
		t.Error("expected a meter")
	}
}

func TestWithExistingBytesMeter(t *testing.T) {
	meter := NewBytesMeter()
	ctx := WithExistingBytesMeter(context.Background(), meter)
	if GetBytesMeter(ctx) != meter {
		t.Error("expected a meter")
	}
}

func TestWithBytesMeter_existing_noop(t *testing.T) {
	ctx := WithExistingBytesMeter(context.Background(), NoopBytesMeter)
	ctx = WithBytesMeter(ctx)

	bm := GetBytesMeter(ctx)
	if bm == NoopBytesMeter {
		t.Error("expected a meter")
	}
}

func TestWithBytesMeter_exists(t *testing.T) {
	meter := NewBytesMeter()
	ctx := WithExistingBytesMeter(context.Background(), meter)
	meter.AddBytesRead(10)

	ctx = WithBytesMeter(ctx)
	bm := GetBytesMeter(ctx)
	if bm == NoopBytesMeter {
		t.Error("expected a meter")
	}
	if bm.BytesRead() != 10 {
		t.Errorf("expected 10, got %d. was a different meter added?", bm.BytesRead())
	}
}

func TestWithExistingBytesMeter_nil(t *testing.T) {
	ctx := WithExistingBytesMeter(context.Background(), nil)
	bm := GetBytesMeter(ctx)
	if bm != NoopBytesMeter {
		t.Error("expected a noop meter")
	}
}
