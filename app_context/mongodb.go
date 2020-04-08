package app_context

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MongodbContextKey = "ctx:mongodb"
)

func WithMongodb(db *mongo.Database) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, MongodbContextKey, db)
	}
}

func GetMongodb(ctx context.Context) *mongo.Database {
	r, ok := ctx.Value(MongodbContextKey).(*mongo.Database)
	if !ok {
		panic(fmt.Sprintf("cannot found key Nats in context"))
	}
	return r
}
