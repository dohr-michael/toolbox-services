package lib

import (
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	ID              string `graphql:"ID"`
	Name            string `graphql:"name"`
	Picture         string `graphql:"picture,optional"`
	ConnectionCount int32  `graphql:"connectionCount"`
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"ID":              &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"name":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"picture":         &graphql.Field{Type: graphql.String},
		"connectionCount": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
	},
})

type Character struct {
	Player    User     `graphql:"player"`
	IsDead    bool     `graphql:"isDead"`
	Abilities []string `graphql:"abilities"`
}

var characterType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Character",
	Fields: graphql.Fields{
		"player":    &graphql.Field{Type: graphql.NewNonNull(userType)},
		"isDead":    &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"abilities": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
	},
})

func TestStructToGraphqlObject(t *testing.T) {
	type args struct {
		item interface{}
		obj  *graphql.Object
	}
	tests := []struct {
		name        string
		configs     []configs
		args        args
		want        *graphql.Object
		expectedErr string
	}{
		{name: "test1", args: args{item: User{}, obj: graphql.NewObject(graphql.ObjectConfig{Name: "User", Fields: graphql.Fields{}}),}, want: userType},
		{name: "test2", args: args{item: &User{}, obj: graphql.NewObject(graphql.ObjectConfig{Name: "User", Fields: graphql.Fields{}}),}, want: nil, expectedErr: graphqlInvalidInput().Error()},
		{name: "test3", args: args{item: Character{}, obj: graphql.NewObject(graphql.ObjectConfig{Name: "Character", Fields: graphql.Fields{}}),}, want: characterType, configs: []configs{WithGraphqlReferences(userType)}},
		{name: "test4", args: args{item: Character{}, obj: graphql.NewObject(graphql.ObjectConfig{Name: "Character", Fields: graphql.Fields{}}),}, want: nil, expectedErr: graphqlRefIsMissing("User").Error()},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := tt.configs
			if conf == nil {
				conf = make([]configs, 0)
			}
			err := StructToGraphqlObject(tt.args.item, tt.args.obj, conf...)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want.Fields(), tt.args.obj.Fields())
			}
		})
	}
}
