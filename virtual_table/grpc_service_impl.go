package virtual_table

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/app_context"
	"github.com/rpg-tools/toolbox-services/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

func NewGrpcService() proto.VirtualTableServiceServer {
	return &grpcServices{
		underlying: &tableServices{},
	}
}

type grpcServices struct {
	underlying *tableServices
}

func (g *grpcServices) CreateTable(context.Context, *proto.CreateTableInput) (*proto.Table, error) {
	panic("implement me")
}

func (g *grpcServices) ListTables(*proto.Void, proto.VirtualTableService_ListTablesServer) error {
	panic("implement me")
}

func (g *grpcServices) GetTable(context.Context, *proto.TableIdInput) (*proto.Table, error) {
	panic("implement me")
}

func (g *grpcServices) JoinTable(context.Context, *proto.JoinTableInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) CreateCharacter(context.Context, *proto.CreateCharacterInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) ChangeCharacterPlayer(context.Context, *proto.ChangeCharacterPlayerInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) CreateDiscussion(context.Context, *proto.CreateDiscussionInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) AddUserToDiscussion(context.Context, *proto.AddUserToDiscussionInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) DeleteDiscussion(context.Context, *proto.DeleteDiscussionInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) SendMessage(context.Context, *proto.SendMessageInput) (*proto.Event, error) {
	panic("implement me")
}

func (g *grpcServices) Subscribe(request *proto.TableIdInput, result proto.VirtualTableService_SubscribeServer) error {
	ctx := result.Context()
	user := app_context.GetAuthUser(ctx)
	natsConn := app_context.GetNats(ctx)
	bsonId, err := primitive.ObjectIDFromHex(request.Id)
	if err != nil {
		log.Printf("bad format for id %v", err)
		return err
	}
	messages := make(chan *nats.Msg)
	// Subscribe to topic
	sub, err := natsConn.ChanSubscribe(request.Id, messages)
	if err != nil {
		log.Printf("cannot subscribe to channed : %v", err)
		return err
	}
	defer sub.Unsubscribe()

	// Notify user connected.
	data, _ := WriteEventJson(&PlayerConnected{
		EventBase: NewEventBase(bsonId, []string{"*"}, user),
		Player:    user,
	})
	err = natsConn.Publish(request.Id, data)
	if err != nil {
		log.Printf("cannot send connected message : %v", err)
		return err
	}

	// Listen messages.
	for {
		select {
		case mess, ok := <-messages:
			if !ok {
				log.Printf("stream %s closed", request.Id)
				return nil
			}
			evt, _ := ReadEventJson(mess.Data)
			log.Printf("send message %v", evt)
			err := result.Send(&proto.Event{Id: evt.GetId(), Type: string(evt.Kind()),})
			if err != nil {
				log.Printf("cannot send message to stream %s", err)
				return err
			}
		case <-ctx.Done():
			log.Printf("player %s disconnected of %s", user, request.Id)
			data, _ := WriteEventJson(&PlayerDisconnected{
				EventBase: NewEventBase(bsonId, []string{"*"}, user),
				Player:    user,
			})
			_ = natsConn.Publish(request.Id, data)
			return nil
		}
	}
}
