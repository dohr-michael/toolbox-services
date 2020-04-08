package backup
/*
import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/lib/graphql_subscription"
	"github.com/rpg-tools/toolbox-services/virtual_table"
)

type ConnectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

func GraphqlRouter(natsConn *nats.Conn) (func(chi.Router), error) {
	natsNotifier := graphql_subscription.NewNatsNotifier(natsConn)
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"table": &graphql.Field{
					Type: tableType,
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
						id, ok := p.Args["id"].(string)
						if ok {
							for _, table := range virtual_table.tables {
								if table.Id == id {
									return table, nil
								}
							}
						}
						return nil, nil
					},
					Description: "",
				},
			},
		}),
		//Mutation:     nil,
		Subscription: graphql.NewObject(graphql.ObjectConfig{
			Name: "Subscriptions",
			Fields: graphql.Fields{
				"connect": &graphql.Field{
					Type: eventType,
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						id, ok := p.Args["id"].(string)
						if !ok {
							return nil, fmt.Errorf("id not found")
						}
						// TODO Check if exits
						err := graphql_subscription.Subscribe(p.Context, id)
						//err := graphql_subscription.Subscribe(p.Context, id)

						return nil, err
					},
				},
			},
			Description: "",
		}),
		Types: []graphql.Type{
			userType,
			characterType,
			discussionType,
			tableType,
			playerConnectedType,
			playerDisconnectedType,
			playerJointType,
			playerLeavedType,
			playerWritingMessageType,
			playerStopWritingMessageType,
			playerSentMessageType,
		},
		//Directives:   nil,
		//Extensions:   nil,
	})
	if err != nil {
		return nil, err
	}

	//upgrader := &websocket.Upgrader{
	//	Subprotocols: []string{"graphql-ws"},
	//	CheckOrigin:  func(r *http.Request) bool { return true },
	//}
	//
	//subHandler := subscriptionHandler(upgrader)
	subHandler := graphql_subscription.NewSubscriptionHandler(natsNotifier, &schema)

	return func(router chi.Router) {
		h := handler.New(&handler.Config{
			Schema:     &schema,
			Pretty:     true,
			Playground: true,
		})
		router.Mount("/subscriptions", subHandler)
		router.Handle("/graphql", h)
	}, nil
}
*/