package file

import (
	"testing"

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
			dsn: "file:///tmp/test?network=testNetwork",
			expect: &Config{
				BasePath:             "/tmp/test",
				Network:              "testNetwork",
				FlushIntervalSeconds: 60,    //default value
				BufferSize:           10000, //default value
			},
		},
		{
			dsn: "file:///tmp/test?network=testNetwork&buffer=25&flushIntervalSeconds=10",
			expect: &Config{
				BasePath:             "/tmp/test",
				Network:              "testNetwork",
				FlushIntervalSeconds: 10,
				BufferSize:           25,
			},
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
