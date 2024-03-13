package context

import "context"

// Add network to context
type networkKeyType string

const networkKey = networkKeyType("network")

func WithNetwork(ctx context.Context, network string) context.Context {
	return context.WithValue(ctx, networkKey, network)
}

func GetNetwork(ctx context.Context) string {
	if network, ok := ctx.Value(networkKey).(string); ok {
		return network
	}
	return ""
}
