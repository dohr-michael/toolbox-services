package lib

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"strings"
)

func graphqlInvalidInput() error              { return fmt.Errorf("expected structur") }
func graphqlInvalidTag(name string) error     { return fmt.Errorf("tag graphql of %s is invalid", name) }
func graphqlRefIsMissing(name string) error   { return fmt.Errorf("reference of %s is missing", name) }
func graphqlTypeNotManaged(name string) error { return fmt.Errorf("type %s not managed as field", name) }

type structToGraphqlContext struct {
	Name             string
	AdditionalFields []*graphql.Field
	References       map[string]*graphql.Object
}

type configs func(*structToGraphqlContext)

func WithGraphqlReferences(references ...*graphql.Object) func(*structToGraphqlContext) {
	return func(ctx *structToGraphqlContext) {
		refs := make(map[string]*graphql.Object)
		for _, ref := range references {
			refs[ref.Name()] = ref
		}
		ctx.References = refs
	}
}

func WithGraphqlFields(fields ...*graphql.Field) func(ctx *structToGraphqlContext) {
	return func(ctx *structToGraphqlContext) {
		ctx.AdditionalFields = fields
	}
}

func tagsContains(tags []string, item string, prefix bool) string {
	for _, tag := range tags {
		trim := strings.TrimSpace(tag)
		if prefix && strings.HasPrefix(trim, item) || !prefix && trim == item {
			return tag
		}
	}
	return ""
}

func toGraphqlType(typ reflect.Type, tags []string, ctx *structToGraphqlContext) (graphql.Type, error) {
	var graphqlTyp graphql.Type
	switch typ.Kind() {
	case reflect.String:
		graphqlTyp = graphql.String
	case reflect.Bool:
		graphqlTyp = graphql.Boolean
	case reflect.Int, reflect.Int32, reflect.Int64:
		graphqlTyp = graphql.Int
	case reflect.Float32, reflect.Float64:
		graphqlTyp = graphql.Float
	case reflect.Ptr:
		return toGraphqlType(typ.Elem(), tags, ctx)
	case reflect.Slice:
		styp, err := toGraphqlType(typ.Elem(), tags, ctx)
		if err != nil {
			return nil, err
		}
		graphqlTyp = graphql.NewList(styp)
	case reflect.Struct:
		ref, ok := ctx.References[typ.Name()]
		if !ok {
			return nil, graphqlRefIsMissing(typ.Name())
		}
		graphqlTyp = ref
	default:
		return nil, graphqlTypeNotManaged(typ.Kind().String())
	}
	if ref := tagsContains(tags, "ref", true); ref != "" {
		refS := strings.Split(ref, ":")
		if len(refS) <= 1 {
			return nil, graphqlRefIsMissing(ref)
		}
		tRef, ok := ctx.References[refS[1]]
		if !ok {
			return nil, graphqlRefIsMissing(refS[1])
		}
		graphqlTyp = tRef
	}
	if tagsContains(tags, "optional", false) == "" {
		graphqlTyp = graphql.NewNonNull(graphqlTyp)
	}
	return graphqlTyp, nil
}

func StructToGraphqlObject(item interface{}, obj *graphql.Object, conf ...configs) error {
	ctx := &structToGraphqlContext{}
	for _, c := range conf {
		c(ctx)
	}
	type tuple struct {
		name string
		tag  string
		t    reflect.Type
	}
	fields := make([]tuple, 0)
	t := reflect.TypeOf(item)
	switch t.Kind() {
	case reflect.Struct:
		if ctx.Name == "" {
			ctx.Name = t.Name()
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag, hasTag := field.Tag.Lookup("graphql")
			if hasTag {
				fields = append(fields, tuple{name: field.Name, tag: tag, t: field.Type})
			}
		}
	default:
		return graphqlInvalidInput()
	}

	for _, f := range fields {
		tagValue := strings.Split(f.tag, ",")
		if len(tagValue) == 0 {
			return graphqlInvalidTag(f.name)
		}
		typ, err := toGraphqlType(f.t, tagValue, ctx)
		if err != nil {
			return err
		}
		obj.AddFieldConfig(tagValue[0], &graphql.Field{Type: typ})
	}
	for _, f := range ctx.AdditionalFields {
		obj.AddFieldConfig(f.Name, f)
	}
	return nil
}
