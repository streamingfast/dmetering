package grpc

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Endpoint    string
	Delay       time.Duration
	BufferSize  uint64
	PanicOnDrop bool
	Network     string
}

func newConfig(configURL string) (*Config, error) {
	c := &Config{
		Delay:       100 * time.Millisecond,
		BufferSize:  10000,
		PanicOnDrop: false,
	}

	u, err := url.Parse(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse urls: %w", err)
	}

	c.Endpoint = u.Host
	if c.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not specified (as hostname)")
	}

	vals := u.Query()
	c.Network = vals.Get("network")
	if c.Network == "" {
		return nil, fmt.Errorf("network not specified (as query param)")
	}

	bufferValue := vals.Get("buffer")
	if bufferValue != "" {
		c.BufferSize, err = strconv.ParseUint(bufferValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid buffer value %q: %w", bufferValue, err)
		}
	}

	delayValue := vals.Get("delay")
	if delayValue != "" {
		delay, err := strconv.ParseInt(delayValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid delay value %q: %w", delayValue, err)
		}

		c.Delay = time.Duration(delay) * time.Millisecond
	}

	c.PanicOnDrop = vals.Get("panicOnDrop") == "true"

	return c, nil
}
