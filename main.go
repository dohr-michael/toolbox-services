//

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/admin"
	"github.com/rpg-tools/toolbox-services/app_context"
	"github.com/rpg-tools/toolbox-services/lib"
	"github.com/rpg-tools/toolbox-services/proto"
	"github.com/rpg-tools/toolbox-services/virtual_table"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func Router(auth0Client string, auth0Secret string, enrichment ...app_context.ContextEnrichment) chi.Router {
	mux := chi.NewMux()

	// A good base middleware stack
	mux.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		middleware.Timeout(60*time.Second),
		lib.AuthHttpMiddleware(auth0Client, auth0Secret, app_context.AuthTokenContextKey),
		app_context.ContextMiddleware(enrichment...),
	)

	mux.Route("/@", admin.Router)
	mux.Route("/virtual-tables", virtual_table.Route)
	return mux
}

func Grpc(enrichment ...app_context.ContextEnrichment) *grpc.Server {
	authAndContextFunction := func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		var username string
		if err != nil && strings.Contains(err.Error(), "Unauthenticated") {
			username = "anonymous"
		} else if err != nil {
			return nil, err
		} else {
			// TODO extract user from jwt token.
			username = "michael"
		}
		log.Printf("token : %s, username : %s", token, username)
		for _, fn := range append(enrichment /*, app_context.WithAuthUser(username), app_context.WithAuthToken(token)*/) {
			ctx = fn(ctx)
		}
		return ctx, nil
	}

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_auth.StreamServerInterceptor(authAndContextFunction),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(authAndContextFunction),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	proto.RegisterVirtualTableServiceServer(server, virtual_table.NewGrpcService())
	return server
}

func main() {
	httpPort := flag.Int("http-port", 8080, "port to bind (default: 8080)")
	grpcPort := flag.Int("grpc-port", 9900, "grpc port to bind (default: 9900)")
	natsUri := flag.String("nats-uri", "127.0.0.1:4222", "nats address (default: 127.0.0.1:4222)")
	mongodbUri := flag.String("mongo-uri", "mongodb://127.0.0.1:27017/rpg-tools", "mongodb address (default: mongodb://127.0.0.1:27017/rpg-tools)")
	auth0ClientId := flag.String("auth0-client-id", "", "auth0 client id")
	auth0ClientSecret := flag.String("auth0-client-secret", "", "auth0 client secret")

	flag.Parse()

	// Init nats
	natsConn, err := nats.Connect(*natsUri)
	if err != nil {
		log.Fatal(err)
	}
	defer natsConn.Close()

	// Init mongodb
	cstring, err := connstring.Parse(*mongodbUri)
	if err != nil {
		log.Fatal(err)
	}
	muri := options.Client().ApplyURI(cstring.String())
	mongoClient, err := mongo.NewClient(muri)
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		_ = mongoClient.Disconnect(ctx)
	}()
	database := mongoClient.Database(cstring.Database)

	// Auth0

	log.Printf("auth0 client : %s, auth0 secret : %s", *auth0ClientId, *auth0ClientSecret)

	enrichments := []app_context.ContextEnrichment{
		app_context.WithNats(natsConn),
		app_context.WithMongodb(database),
	}
	// Init router
	router := Router(*auth0ClientId, *auth0ClientSecret, enrichments...)

	// Grpc server
	grpcServer := Grpc(enrichments...)

	errors := make(chan error, 1)

	go func() {
		log.Printf("start http server, listen :%d", *httpPort)
		errors <- http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), router)
	}()
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			errors <- err
			return
		}

		log.Printf("start grpc server, listen :%d", *grpcPort)
		errors <- grpcServer.Serve(lis)
	}()

	err = <-errors

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("server stopped")
	os.Exit(0)
}
