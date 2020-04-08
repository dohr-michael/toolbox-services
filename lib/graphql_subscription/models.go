package graphql_subscription

import "encoding/json"

type payload struct {
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
	Extensions    interface{}            `json:"extensions"`
	Query         string                 `json:"query"`
}

type wsRequest struct {
	Id      string  `json:"id"`
	Type    string  `json:"type"`
	Payload payload `json:"payload"`
}

type wsMessage struct {
	Id      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (m wsMessage) ToGraphQLRequest() (*wsRequest, error) {
	q := &wsRequest{
		Type: m.Type,
		Id:   m.Id,
	}

	d, err := json.Marshal(m.Payload)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(d, &q.Payload)
	if err != nil {
		return nil, err
	}

	return q, nil
}
