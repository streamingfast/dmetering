package dmetering

import "testing"

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
