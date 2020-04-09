package virtual_table

import (
	"encoding/json"
	"fmt"
	"github.com/rpg-tools/toolbox-services/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type EventType string

const (
	TableCreatedType             EventType = "evt:table-created"
	PlayerJointType              EventType = "evt:player-joint"
	PlayerConnectedType          EventType = "evt:player-connected"
	PlayerDisconnectedType       EventType = "evt:player-disconnected"
	PlayerWritingMessageType     EventType = "evt:player-writing-message"
	PlayerStopWritingMessageType EventType = "evt:player-stop-writing-message"
	PlayerSentMessageType        EventType = "evt:player-sent-message"
)

var eventsSupplierByKind map[EventType]func() Event

func init() {
	eventsSupplierByKind = map[EventType]func() Event{
		TableCreatedType:             func() Event { return &TableCreated{} },
		PlayerJointType:              func() Event { return &PlayerJoint{} },
		PlayerConnectedType:          func() Event { return &PlayerConnected{} },
		PlayerDisconnectedType:       func() Event { return &PlayerDisconnected{} },
		PlayerWritingMessageType:     func() Event { return &PlayerWritingMessage{} },
		PlayerStopWritingMessageType: func() Event { return &PlayerStopWritingMessage{} },
		PlayerSentMessageType:        func() Event { return &PlayerSentMessage{} },
	}
}

type Event interface {
	GetAt() time.Time
	GetBy() string
	GetId() string
	GetTableId() string
	GetAllowUsers() []string
	Kind() EventType
}

type EventBase struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	TableId    primitive.ObjectID `json:"tableId" bson:"tableId"`
	By         string             `json:"by" bson:"_by"`
	AllowUsers []string           `json:"-" bson:"allowUsers"`
}

func (e *EventBase) GetId() string           { return e.Id.Hex() }
func (e *EventBase) GetTableId() string      { return e.TableId.Hex() }
func (e *EventBase) GetAt() time.Time        { return e.Id.Timestamp() }
func (e *EventBase) GetAllowUsers() []string { return e.AllowUsers }
func (e *EventBase) GetBy() string           { return e.By }

func NewEventBase(tableId primitive.ObjectID, allowUsers []string, by string) EventBase {
	return EventBase{
		Id:         primitive.NewObjectID(),
		TableId:    tableId,
		AllowUsers: allowUsers,
		By:         by,
	}
}

func WriteEvent(evt Event, tagName string) (map[string]interface{}, error) {
	res, err := lib.AsMap(evt, tagName)
	if err != nil {
		return nil, err
	}
	res["_kind"] = evt.Kind()
	return res, nil
}

func WriteEventJson(evt Event) ([]byte, error) {
	m, err := WriteEvent(evt, "json")
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func WriteEventBson(evt Event) ([]byte, error) {
	m, err := WriteEvent(evt, "bson")
	if err != nil {
		return nil, err
	}
	return bson.Marshal(m)
}

func ReadEvent(data map[string]interface{}, tagName string) (Event, error) {
	kind, ok := data["_kind"]
	if !ok {
		return nil, fmt.Errorf("'_kind' not found in data")
	}
	if v, b := kind.(string); b {
		sup, ok := eventsSupplierByKind[EventType(v)]
		if !ok || sup == nil {
			return nil, fmt.Errorf("%s is not a valid event", v)
		}
		evt := sup()
		err := lib.FromMap(data, evt, tagName)
		if err != nil {
			return nil, err
		}
		return evt, nil
	}
	return nil, fmt.Errorf("'_kind' is not a string")
}

func ReadEventJson(data []byte) (Event, error) {
	value := make(map[string]interface{})
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	return ReadEvent(value, "json")
}

func ReadEventBson(data []byte) (Event, error) {
	value := make(map[string]interface{})
	err := bson.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	return ReadEvent(value, "bson")
}

// Events

type TableCreated struct {
	EventBase
	Name string `json:"name" bson:"name"`
}

func (*TableCreated) Kind() EventType                { return TableCreatedType }
func (e *TableCreated) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *TableCreated) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerJoint struct {
	EventBase
	Player string `json:"player" bson:"player"`
}

func (*PlayerJoint) Kind() EventType                { return PlayerJointType }
func (e *PlayerJoint) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerJoint) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerConnected struct {
	EventBase
	Player string `json:"player" bson:"player"`
}

func (*PlayerConnected) Kind() EventType                { return PlayerConnectedType }
func (e *PlayerConnected) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerConnected) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerDisconnected struct {
	EventBase
	Player string `json:"player" bson:"player"`
}

func (*PlayerDisconnected) Kind() EventType                { return PlayerDisconnectedType }
func (e *PlayerDisconnected) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerDisconnected) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerWritingMessage struct {
	EventBase
	Player     string `json:"player" bson:"player"`
	Discussion string `json:"discussion" bson:"discussion"`
}

func (*PlayerWritingMessage) Kind() EventType                { return PlayerWritingMessageType }
func (e *PlayerWritingMessage) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerWritingMessage) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerStopWritingMessage struct {
	EventBase
	Player     string `json:"player" bson:"player"`
	Discussion string `json:"discussion" bson:"discussion"`
}

func (*PlayerStopWritingMessage) Kind() EventType                { return PlayerStopWritingMessageType }
func (e *PlayerStopWritingMessage) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerStopWritingMessage) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }

type PlayerSentMessage struct {
	EventBase
	Player     string `json:"player" bson:"player"`
	Discussion string `json:"discussion" bson:"discussion"`
	Message    string `json:"message" bson:"message"`
}

func (*PlayerSentMessage) Kind() EventType                { return PlayerSentMessageType }
func (e *PlayerSentMessage) MarshalBSON() ([]byte, error) { return WriteEventBson(e) }
func (e *PlayerSentMessage) MarshalJSON() ([]byte, error) { return WriteEventJson(e) }
