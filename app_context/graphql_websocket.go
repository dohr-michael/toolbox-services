package app_context

import (
	"context"
	"github.com/rpg-tools/toolbox-services/lib/graphql_subscription"
)

func GetGraphQLWebsocketId(ctx context.Context) string {
	return graphql_subscription.GetWebsocketId(ctx)
}

func GetGraphQLWebsocketSubscribe(ctx context.Context) func(string) {
	return graphql_subscription.GetWebsocketSubscribe(ctx)
}
