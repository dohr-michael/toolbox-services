package backup

import (
	"github.com/graphql-go/graphql"
	"github.com/rpg-tools/toolbox-services/lib"
	"github.com/rpg-tools/toolbox-services/virtual_table"
	"log"
)

var (
	eventType = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Event",
		Fields: graphql.Fields{
			"_kind": &graphql.Field{Type: graphql.String,},
		},
	})
	playerJointType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerJoint",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerJoint, *virtual_table.PlayerJoint:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
	playerConnectedType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerConnected",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerConnected, *virtual_table.PlayerConnected:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
	playerDisconnectedType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerDisconnected",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerDisconnected, *virtual_table.PlayerDisconnected:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
	playerWritingMessageType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerWritingMessage",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerWritingMessage, *virtual_table.PlayerWritingMessage:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
	playerStopWritingMessageType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerStopWritingMessage",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerStopWritingMessage, *virtual_table.PlayerStopWritingMessage:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
	playerSentMessageType = graphql.NewObject(graphql.ObjectConfig{
		Name:       "PlayerSentMessage",
		Interfaces: []*graphql.Interface{eventType},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			switch p.Value.(type) {
			case virtual_table.PlayerSentMessage, *virtual_table.PlayerSentMessage:
				return true
			default:
				return false
			}
		},
		Fields: graphql.Fields{},
	})
)

func init() {
	references := lib.WithGraphqlReferences(userType, characterType, discussionType, tableType, )
	addFields := lib.WithGraphqlFields(&graphql.Field{
		Name: "_kind",
		Type: graphql.NewNonNull(graphql.String),
		Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
			evt, ok := p.Source.(virtual_table.Event)
			if !ok {
				return nil, nil
			}
			return evt.GetKind(), nil
		},
	})
	err := lib.Validate(
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerJoint{}, playerJointType, references, addFields) },
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerConnected{}, playerConnectedType, references, addFields) },
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerDisconnected{}, playerDisconnectedType, references, addFields) },
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerWritingMessage{}, playerWritingMessageType, references, addFields) },
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerStopWritingMessage{}, playerStopWritingMessageType, references, addFields) },
		func() error { return lib.StructToGraphqlObject(virtual_table.PlayerSentMessage{}, playerSentMessageType, references, addFields) },
	)
	if err != nil {
		log.Fatal(err)
	}
}
