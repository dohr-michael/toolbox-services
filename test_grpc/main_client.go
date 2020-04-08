package main

import (
	"context"
	"github.com/rpg-tools/toolbox-services/proto"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:9900", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("cannot connect to grpc service %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := proto.NewVirtualTableServiceClient(conn)

	stream, err := client.Subscribe(context.Background(), &proto.TableIdInput{Id: "t1"})
	if err != nil {
		log.Fatalf("cannot list events %v", err)
	}
	waitstream := make(chan struct{})
	go func() {
		for {
			item, err := stream.Recv()
			if err == io.EOF {
				// read done
				close(waitstream)
				break
			}
			if err != nil {
				log.Fatalf("Failed to receive message : %v", err)
			}
			log.Printf("Message %v", item)
		}
	}()
	<-waitstream
}
