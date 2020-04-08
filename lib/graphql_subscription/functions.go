package graphql_subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

func sendWSMessage(socket *websocket.Conn, msg wsMessage) error {
	data, _ := json.Marshal(msg)
	return socket.WriteMessage(websocket.TextMessage, data)
}

// Subscribe calls a subscription function from handler that has been passed through context
func Subscribe(ctx context.Context, topic string) error {
	mci := ctx.Value(WebsocketSubscribeContextKey)
	if mci != nil {
		mci.(func(string))(topic)
		return nil
	}

	return fmt.Errorf("no subscribe function in context")
}

func withWebsocketId(id string, ctx context.Context) context.Context {
	return context.WithValue(ctx, WebsocketIdContextKey, id)
}

func GetWebsocketId(ctx context.Context) string {
	r, ok := ctx.Value(WebsocketIdContextKey).(string)
	if !ok {
		panic(fmt.Sprintf("cannot found key WebsocketId in context"))
	}
	return r
}

func withWebsocketSubscribe(fn func(string), ctx context.Context) context.Context {
	return context.WithValue(ctx, WebsocketSubscribeContextKey, fn)
}

func GetWebsocketSubscribe(ctx context.Context) func(string) {
	r, ok := ctx.Value(WebsocketSubscribeContextKey).(func(string))
	if !ok {
		panic(fmt.Sprintf("cannot found key WebsocketId in context"))
	}
	return r
}
