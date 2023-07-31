package grpc

import (
	"fmt"
	"net/url"
	"strconv"
)

type Config struct {
	Endpoint    string
	BatchSize   uint64
	BufferSize  uint64
	PanicOnDrop bool
	Network     string
}

func newConfig(configURL string) (*Config, error) {
	c := &Config{
		BatchSize:   100,
		BufferSize:  1000,
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

	if vals.Get("buffer") != "" {
		c.BufferSize, err = strconv.ParseUint(vals.Get("buffer"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid buffer value: %w", err)
		}
	}

	c.PanicOnDrop = vals.Get("panicOnDrop") == "true"

	return c, nil
}
