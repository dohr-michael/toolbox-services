package graphql_subscription

type NotifierCallback func(map[string]interface{})

// Notifier is a interface used for pub-sub like interactions
type Notifier interface {
	// Subscribes the specified function from topic
	Subscribe(topic string, cb NotifierCallback)

	// Unsubscribes the specified function from topic
	Unsubscribe(topic string, cb NotifierCallback)

	// Notify notifies the topic with the optional data
	Notify(topic string, data map[string]interface{})
}
