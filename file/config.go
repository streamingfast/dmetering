package file

import (
	"fmt"
	"net/url"
	"strconv"
)

type Config struct {
	Network              string
	BasePath             string
	BufferSize           uint64
	FlushIntervalSeconds int
	Source               string
}

func newConfig(configURL string) (*Config, error) {
	c := &Config{
		BufferSize:           10000,
		FlushIntervalSeconds: 60,
	}

	u, err := url.Parse(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse urls: %w", err)
	}

	vals := u.Query()

	bufferValue := vals.Get("buffer")
	if bufferValue != "" {
		c.BufferSize, err = strconv.ParseUint(bufferValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid buffer value %q: %w", bufferValue, err)
		}
	}

	rolloverValue := vals.Get("flushIntervalSeconds")
	if rolloverValue != "" {
		rolloverInt, err := strconv.Atoi(rolloverValue)
		if err != nil {
			return nil, fmt.Errorf("invalid rollover value %q: %w", rolloverValue, err)
		}
		c.FlushIntervalSeconds = rolloverInt
	}

	c.BasePath = u.Path
	if c.BasePath == "" {
		return nil, fmt.Errorf("endpoint not specified (as hostname)")
	}

	c.Network = vals.Get("network")

	c.Source = vals.Get("source")

	return c, nil
}
