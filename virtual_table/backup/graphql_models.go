package backup

import (
	"github.com/graphql-go/graphql"
	"github.com/rpg-tools/toolbox-services/lib"
	"github.com/rpg-tools/toolbox-services/virtual_table"
	"log"
)

var (
	userType       = graphql.NewObject(graphql.ObjectConfig{Name: "User", Fields: graphql.Fields{},})
	characterType  = graphql.NewObject(graphql.ObjectConfig{Name: "Character", Fields: graphql.Fields{},})
	discussionType = graphql.NewObject(graphql.ObjectConfig{Name: "Discussion", Fields: graphql.Fields{},})
	tableType      = graphql.NewObject(graphql.ObjectConfig{Name: "Table", Fields: graphql.Fields{},})
)

func init() {
	references := lib.WithGraphqlReferences(userType, characterType, discussionType, tableType, )
	err := lib.Validate(
		func() error { return lib.StructToGraphqlObject(virtual_table.User{}, userType, references) },
		func() error { return lib.StructToGraphqlObject(virtual_table.Character{}, characterType, references) },
		func() error { return lib.StructToGraphqlObject(virtual_table.Discussion{}, discussionType, references) },
		func() error { return lib.StructToGraphqlObject(virtual_table.Table{}, tableType, references) },
	)
	if err != nil {
		log.Fatal(err)
	}
}
