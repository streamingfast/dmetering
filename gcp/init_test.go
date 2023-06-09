package metering_gcp

import (
	"github.com/streamingfast/logging"
)

var zlog, _ = logging.PackageLogger("lidar", "github.com/streamingfast/dmetring/gcp.test")

func init() {
	logging.InstantiateLoggers()
}
