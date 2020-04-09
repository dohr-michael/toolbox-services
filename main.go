//

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/admin"
	"github.com/rpg-tools/toolbox-services/api/virtual_table"
	"github.com/rpg-tools/toolbox-services/app_context"
	"github.com/rpg-tools/toolbox-services/lib"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"log"
	"net/http"
	"os"
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

func main() {
	httpPort := flag.Int("http-port", 8080, "port to bind (default: 8080)")
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

	errors := make(chan error, 1)

	go func() {
		log.Printf("start http server, listen :%d", *httpPort)
		errors <- http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), router)
	}()

	err = <-errors

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("server stopped")
	os.Exit(0)
}
