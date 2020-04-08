package graphql_subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	"net/http"
	"sync"
	"time"
)

type SubscriptionHandler struct {
	upgrader websocket.Upgrader
	notifier Notifier
	schema   *graphql.Schema
}

func NewSubscriptionHandler(
	notifier Notifier,
	schema *graphql.Schema,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		notifier: notifier,
		schema:   schema,
		upgrader: websocket.Upgrader{
			Subprotocols: []string{"graphql-ws"},
			CheckOrigin:  func(r *http.Request) bool { return true },
		},
	}
}

func (sh *SubscriptionHandler) SetCheckOrigin(fn func(r *http.Request) bool) {
	sh.upgrader.CheckOrigin = fn
}

func (sh *SubscriptionHandler) dryRunQuery(ctx context.Context, socket *websocket.Conn, queryData *wsRequest) bool {
	res := graphql.Do(graphql.Params{
		Schema:         *sh.schema,
		RequestString:  queryData.Payload.Query,
		VariableValues: queryData.Payload.Variables,
		OperationName:  queryData.Payload.OperationName,
		Context:        ctx,
	})

	return len(res.Errors) == 0
}

func (sh *SubscriptionHandler) runQuery(ctx context.Context, socket *websocket.Conn, queryData *wsRequest, rootObject map[string]interface{}) bool {
	res := graphql.Do(graphql.Params{
		Schema:         *sh.schema,
		RequestString:  queryData.Payload.Query,
		VariableValues: queryData.Payload.Variables,
		OperationName:  queryData.Payload.OperationName,
		Context:        ctx,
		RootObject:     rootObject,
	})

	err := sendWSMessage(socket, wsMessage{
		Id:      queryData.Id,
		Type:    "data",
		Payload: res,
	})

	return len(res.Errors) == 0 && err == nil
}

func (sh *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try upgrade connection to Websocket
	socket, err := sh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("Error upgrading connection: %s", err)))
		return
	}
	defer func() { _ = socket.Close() }()

	// Ack connection
	//connectionACK, err := json.Marshal(map[string]string{
	//	"type": "connection_ack",
	//})
	//if err != nil {
	//	log.Printf("failed to marshal ws connection ack: %v", err)
	//}
	//if err := socket.WriteMessage(websocket.TextMessage, connectionACK); err != nil {
	//	log.Printf("failed to write to ws connection: %v", err)
	//	return
	//}

	// Create new UUID for this websocket connection
	wsId := uuid.New().String()

	// Inject on Context
	ctx := withWebsocketId(wsId, r.Context())

	// On Change Callbacks
	onChangeChannel := make(chan map[string]interface{})
	onChange := func(data map[string]interface{}) {
		onChangeChannel <- data
	}

	// Track Subscribed Topics
	topics := make(map[string]string, 0)
	l := sync.Mutex{}

	// Create the subscribe function for the query resolvers
	subTopic := func(topic string) {
		l.Lock()
		if _, ok := topics[topic]; !ok {
			if sh.notifier != nil {
				sh.notifier.Subscribe(topic, onChange)
			}
			topics[topic] = topic
		}
		l.Unlock()
	}

	// Inject subscribe function in query resolver
	ctx = withWebsocketSubscribe(subTopic, ctx)

	// Create a callback channel for graphql websocket messages
	newMessage := make(chan wsMessage)

	// Start the clientLoop in a async goroutine
	go func() { _ = sh.clientLoop(socket, newMessage) }()

	// Wait for the message with type `start`
	var msg wsMessage

	msg.Type = "nil"

	select {
	case msg = <-newMessage:
	case <-time.After(startMessageTimeout): // Timeout
	}

	if msg.Type == "nil" {
		return
	}

	m, err := msg.ToGraphQLRequest()

	if err != nil {
		// Timed out
		return
	}

	// Run the first query
	subRunning := sh.dryRunQuery(ctx, socket, m)

	// Loop until a stop message or disconnect is received
	for subRunning {
		select {
		case data := <-onChangeChannel:
			subRunning = sh.runQuery(ctx, socket, m, data)
		case msg := <-newMessage:
			if msg.Type == "stop" || msg.Type == "disconnected" {
				subRunning = false
				break
			}
		}
	}

	// Clean-up all subscribed topics
	if sh.notifier != nil {
		for _, topic := range topics {
			sh.notifier.Unsubscribe(topic, onChange)
		}
	}
}

func (sh *SubscriptionHandler) clientLoop(socket *websocket.Conn, newMessage chan wsMessage) error {
	for {
		_, msgBytes, err := socket.ReadMessage()
		if err != nil {
			newMessage <- wsMessage{
				Type: "disconnected",
			}
			return err
		}

		queryData := wsMessage{}

		err = json.Unmarshal(msgBytes, &queryData)

		if err != nil {
			newMessage <- wsMessage{
				Type:    "error",
				Payload: err.Error(),
			}
			return err
		}

		if queryData.Type == "start" || queryData.Type == "stop" {
			// Only notify start/stop
			newMessage <- queryData
		}

		if queryData.Type == "stop" {
			// Break this loop
			return nil
		}
	}
}
