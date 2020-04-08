package graphql_subscription

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"reflect"
	"sync"
)

type natsEventHandler struct {
	subscription *nats.Subscription
	callback     reflect.Value
}

type natsNotifier struct {
	conn     *nats.Conn
	mutex    sync.Mutex
	handlers map[string][]*natsEventHandler
}

func NewNatsNotifier(conn *nats.Conn) Notifier {
	return &natsNotifier{
		conn:     conn,
		mutex:    sync.Mutex{},
		handlers: make(map[string][]*natsEventHandler),
	}
}

func (n *natsNotifier) Subscribe(topic string, cb NotifierCallback) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	value := reflect.ValueOf(cb)
	handlers, ok := n.handlers[topic]
	if !ok {
		handlers = make([]*natsEventHandler, 0)
	}
	for _, h := range handlers {
		if h.callback == value {
			return
		}
	}
	sub, err := n.conn.Subscribe(topic, func(msg *nats.Msg) {
		res := make(map[string]interface{})
		err := json.Unmarshal(msg.Data, res)
		// TODO Manage errors
		if err != nil {
			cb(res)
		}
	})
	if err == nil {
		n.handlers[topic] = append(handlers, &natsEventHandler{sub, value})
	}
}

func (n *natsNotifier) Unsubscribe(topic string, cb NotifierCallback) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	value := reflect.ValueOf(cb)
	handlers, ok := n.handlers[topic]
	if !ok {
		handlers = make([]*natsEventHandler, 0)
	}
	newHandlers := make([]*natsEventHandler, 0)
	for _, c := range handlers {
		if c.callback == value {
			// TODO Manage errors
			_ = c.subscription.Unsubscribe()
		} else {
			newHandlers = append(newHandlers, c)
		}
	}
	n.handlers[topic] = newHandlers
}

func (n *natsNotifier) Notify(topic string, data map[string]interface{}) {
	// TODO Manage error
	res, err := json.Marshal(data)
	if err != nil {
		_ = n.conn.Publish(topic, res)
	}
}
