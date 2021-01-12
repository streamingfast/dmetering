package dmetering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZlogPlugin_ParseDSN(t *testing.T) {
	tests := []struct {
		name              string
		in                string
		expectedPackageID string
		expectedLevel     zapcore.Level
		expectedErr       error
	}{
		{"success, no package id", "zlog://", defaultPackageID, zap.InfoLevel, nil},
		{"success, with level", "zlog://?level=warn", defaultPackageID, zap.WarnLevel, nil},
		{"success, with package id", "zlog://test.com/value", "test.com/value", zap.InfoLevel, nil},
		{"success, with package id & level", "zlog://test.com/value?level=warn", "test.com/value", zap.WarnLevel, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packageID, level, err := parseZlogPluginDSN(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expectedPackageID, packageID)
				assert.Equal(t, test.expectedLevel, level)

			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}
