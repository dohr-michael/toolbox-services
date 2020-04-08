package app_context

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
)

const (
	NatsContextKey = "ctx:nats"
)

func WithNats(conn *nats.Conn) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, NatsContextKey, conn)
	}
}

func GetNats(ctx context.Context) *nats.Conn {
	r, ok := ctx.Value(NatsContextKey).(*nats.Conn)
	if !ok {
		panic(fmt.Sprintf("cannot found key Nats in context"))
	}
	return r
}
