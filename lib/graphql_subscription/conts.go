package graphql_subscription

import (
	"time"
)

const (
	WebsocketIdContextKey        = "ctx:websocket:id"
	WebsocketSubscribeContextKey = "ctx:websocket:subscribe"
	startMessageTimeout          = 2 * time.Second
)
