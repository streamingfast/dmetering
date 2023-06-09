package metering_gcp

import (
	"context"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func NewFakePubsub(ctx context.Context) (*pubsub.Client, *pstest.Server) {
	// Start a fake server running locally.
	srv := pstest.NewServer()
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, "P", option.WithGRPCConn(conn))
	if err != nil {
		panic(err)
	}
	return client, srv
}
