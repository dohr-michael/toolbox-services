package backup

/*
import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/lib/subscription"
	"github.com/rpg-tools/toolbox-services/virtual_table"
)

type natsSubscribeService struct {
	natsConn *nats.Conn
	schema   graphql.Schema
}

func newSubscribeService(natsConn *nats.Conn, schema graphql.Schema) subscription.SubscribeService {
	return &natsSubscribeService{
		natsConn: natsConn,
		schema:   schema,
	}
}

func (n *natsSubscribeService) getTopicName(
	document string,
	operationName string,
	variableValues map[string]interface{},
) (string, error) {
	idI, ok := variableValues["id"]
	if !ok {
		return "", fmt.Errorf("variable 'id' not found")
	}
	id, ok := idI.(string)
	if !ok {
		return "", fmt.Errorf("'id' is not a string")
	}
	return id, nil
}

func (n *natsSubscribeService) sendConnectedEvent(
	ctx context.Context,
	document string,
	operationName string,
	variableValues map[string]interface{},
) {
	n.sendEvent(
		&virtual_table.PlayerConnected{uuid.New().String(), "michael"},
		document,
		operationName,
		variableValues,
	)
}

func (n *natsSubscribeService) sendDisconnectedEvent(
	ctx context.Context,
	document string,
	operationName string,
	variableValues map[string]interface{},
) {
	n.sendEvent(
		&virtual_table.PlayerDisconnected{uuid.New().String(), "michael"},
		document,
		operationName,
		variableValues,
	)
}

func (n *natsSubscribeService) sendEvent(
	evt virtual_table.Event,
	document string,
	operationName string,
	variableValues map[string]interface{},
) {
	topic, _ := n.getTopicName(document, operationName, variableValues)
	data, _ := virtual_table.WriteEventJson(evt)
	_ = n.natsConn.Publish(topic, data)
}

func (n *natsSubscribeService) Subscribe(
	ctx context.Context,
	document string,
	operationName string,
	variableValues map[string]interface{},
) (<-chan interface{}, func(), []gqlerrors.FormattedError) {
	gqlResult := graphql.Do(graphql.Params{
		Schema:         n.schema,
		RequestString:  document,
		VariableValues: variableValues,
		OperationName:  operationName,
		Context:        ctx,
	})
	if gqlResult.HasErrors() {
		return nil, nil, gqlResult.Errors
	}

	data := make(chan *nats.Msg)
	topicName, err := n.getTopicName(document, operationName, variableValues)
	if err != nil {
		return nil, nil, gqlerrors.FormatErrors(err)
	}
	sub, err := n.natsConn.ChanSubscribe(topicName, data)
	if err != nil {
		return nil, nil, gqlerrors.FormatErrors(err)
	}
	result := make(chan interface{})
	go func() {
		for {
			d, ok := <-data
			if !ok {
				close(result)
				return
			}
			// TODO
			result <- d.Data
		}
	}()
	n.sendConnectedEvent(ctx, document, operationName, variableValues)
	return result, func() {
		_ = sub.Unsubscribe()
		n.sendDisconnectedEvent(ctx, document, operationName, variableValues)
	}, nil
}
*/
