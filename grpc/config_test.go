package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_new(t *testing.T) {
	tests := []struct {
		dsn         string
		expect      *Config
		expectError bool
	}{
		{
			dsn: "grpc://localhost:9010?buffer=25&network=eth-mainnet",
			expect: &Config{
				Endpoint:   "localhost:9010",
				Network:    "eth-mainnet",
				Delay:      100 * time.Millisecond,
				BufferSize: 25,
			},
		},
		{
			dsn: "grpc://localhost:9010?buffer=100000&network=eth-mainnet&panicOnDrop=true",
			expect: &Config{
				Endpoint:    "localhost:9010",
				Network:     "eth-mainnet",
				Delay:       100 * time.Millisecond,
				BufferSize:  100000,
				PanicOnDrop: true,
			},
		},
		{
			dsn: "grpc://localhost:9010?buffer=100000&network=eth-mainnet&delay=250",
			expect: &Config{
				Endpoint:   "localhost:9010",
				Network:    "eth-mainnet",
				Delay:      250 * time.Millisecond,
				BufferSize: 100000,
			},
		},
		{
			dsn:         "grpc:localhost9010?buffer=100000&network=eth-mainnet&panicOnDrop=true",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.dsn, func(t *testing.T) {
			c, err := newConfig(test.dsn)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expect, c)
			}
		})
	}
}
