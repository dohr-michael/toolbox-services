package virtual_table

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rpg-tools/toolbox-services/app_context"
	"github.com/rpg-tools/toolbox-services/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionName        = "tables"
	journalCollectionName = "tables_journal"
)

type tableServices struct{}

func (*tableServices) withMongoTransaction(ctx context.Context, db *mongo.Database, fn func() error) error {
	s, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	err = s.StartTransaction()
	defer s.EndSession(ctx)
	if err != nil {
		return err
	}
	return mongo.WithSession(ctx, s, func(sc mongo.SessionContext) error {
		if err := fn(); err != nil {
			return err
		}
		if err := s.CommitTransaction(sc); err != nil {
			return err
		}
		return nil
	})
}

func (*tableServices) sendEvent(ctx context.Context, table primitive.ObjectID, evt Event) {
	natsConn := app_context.GetNats(ctx)
	p, err := json.Marshal(evt)
	if err != nil {
		// TODO manage error
		return
	}
	_ = natsConn.Publish(table.Hex(), p)
}

// Read part

func (*tableServices) aggregate(filters bson.M, limit int, ctx context.Context) ([]*TableWithEvents, error) {
	db := app_context.GetMongodb(ctx)
	user := app_context.GetAuthUser(ctx)

	pipelines := mongo.Pipeline{
		bson.D{{"$match", filters}},
		bson.D{{"$unwind", bson.M{"path": "$discussions", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$match", bson.M{"$or": bson.A{
			bson.M{"discussions.between": user},
			bson.M{"discussions.between": "*"},
		}}}},
		bson.D{{"$group", bson.M{"_id": "$_id", "doc": bson.M{"$first": "$$ROOT"}, "discussions": bson.M{"$push": "$discussions"}}}},
		bson.D{{"$replaceRoot", bson.M{"newRoot": bson.M{"$mergeObjects": bson.A{"$doc", bson.M{"discussions": "$discussions"}}}}}},
		bson.D{{"$lookup", bson.M{
			"from": journalCollectionName,
			"let": bson.M{
				"tableId": "$_id",
			},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$tableId", "$$tableId"}},
							bson.M{"$or": bson.A{
								bson.M{"$in": bson.A{"*", "$allowUsers"}},
								bson.M{"$in": bson.A{user, "$allowUsers"}},
							}},
						},
					},},
				},
			},
			"as": "events",
		}}},
	}
	if limit > 0 {
		pipelines = append(pipelines, bson.D{{"$limit", limit}})
	}

	cursor, err := db.Collection(collectionName).Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	type aggregateResult struct {
		Events []map[string]interface{} `bson:"events"`
		Table  `bson:",inline"`
	}
	all := make([]aggregateResult, 0)
	if err = cursor.All(ctx, &all); err != nil {
		return nil, err
	}
	res := make([]*TableWithEvents, len(all))
	for idx, item := range all {
		obj := &TableWithEvents{Table: item.Table, Events: make([]Event, 0)}
		// Read events
		for _, evt := range item.Events {
			e, err := ReadEvent(evt, "bson")
			if err != nil {
				return nil, err
			}
			obj.Events = append(obj.Events, e)
		}
		res[idx] = obj
	}
	return res, nil
}

// TODO Indexation
// TODO As stream
func (s *tableServices) Search(filters interface{}, ctx context.Context) ([]*TableWithEvents, error) {
	// TODO Filters...
	return s.aggregate(bson.M{}, -1, ctx)
}

func (s *tableServices) ById(id string, ctx context.Context) (*TableWithEvents, error) {
	bsonId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	res, err := s.aggregate(bson.M{"_id": bsonId}, 1, ctx)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

// Commands.

func (s *tableServices) CreateTable(cmd CreateTableCmd, ctx context.Context) (Event, error) {
	db := app_context.GetMongodb(ctx)
	user := app_context.GetAuthUser(ctx)

	table := Table{Id: primitive.NewObjectID(), Name: cmd.Name, Master: user, Players: []string{}, Characters: []Character{}, Discussions: []Discussion{
		{Id: uuid.New().String(), Name: "General", Persistent: true, Between: []string{"*"}, Messages: []Message{}},
	}}
	evt := TableCreated{EventBase: NewEventBase(table.Id, []string{"*"}, user), Name: cmd.Name}
	evtAsMap, err := WriteEvent(&evt, "bson")
	if err != nil {
		return nil, err
	}
	tableAsMap, err := lib.AsMap(&table, "bson")
	if err != nil {
		return nil, err
	}
	if err := s.withMongoTransaction(ctx, db, func() error {
		if _, err := db.Collection(journalCollectionName).InsertOne(ctx, evtAsMap); err != nil {
			return err
		}
		if _, err := db.Collection(collectionName).InsertOne(ctx, tableAsMap); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	s.sendEvent(ctx, table.Id, &evt)
	return &evt, nil
}
