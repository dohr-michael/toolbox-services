package virtual_table

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/rpg-tools/toolbox-services/app_context"
	"github.com/rpg-tools/toolbox-services/lib"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 512
)

func createTableRoute(services *tableServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := json.NewDecoder(r.Body)
		payload := CreateTableCmd{}
		if err := d.Decode(&payload); err != nil {
			_ = render.Render(w, r, lib.HttpBadRequest(err))
			return
		}
		res, err := services.CreateTable(payload, r.Context())
		if err != nil {
			_ = render.Render(w, r, lib.ToHttpError(err))
			return
		}
		if err = render.Render(w, r, lib.HttpResponseWithId(res.GetTableId(), res, http.StatusCreated)); err != nil {
			_ = render.Render(w, r, lib.HttpRenderError(err))
			return
		}
	}
}

func findManyTableRoute(services *tableServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO manage query filter.
		ctx := r.Context()
		tables, err := services.Search(nil, ctx)
		if err != nil {
			_ = render.Render(w, r, lib.ToHttpError(err))
			return
		}
		if err = render.Render(w, r, lib.HttpResponse(tables, 200)); err != nil {
			_ = render.Render(w, r, lib.HttpRenderError(err))
			return
		}
	}
}

func findOneTableRoute(services *tableServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		table, err := services.ById(id, ctx)
		if err != nil {
			_ = render.Render(w, r, lib.ToHttpError(err))
			return
		}
		if table == nil {
			_ = render.Render(w, r, lib.HttpNotFound(nil))
			return
		}
		if err = render.Render(w, r, lib.HttpResponse(table, 200)); err != nil {
			_ = render.Render(w, r, lib.HttpRenderError(err))
			return
		}
	}
}

func Route(router chi.Router) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	services := &tableServices{}

	router.Post("/", createTableRoute(services))
	router.Get("/", findManyTableRoute(services))
	router.Get("/{id}", findOneTableRoute(services))

	router.Mount("/{id}/subscribe", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		natsConn := app_context.GetNats(ctx)
		user := app_context.GetAuthUser(ctx)
		id := chi.URLParam(r, "id")
		table, err := services.ById(id, ctx)
		if err != nil {
			_ = render.Render(w, r, lib.ToHttpError(err))
			return
		}
		if table == nil {
			_ = render.Render(w, r, lib.HttpNotFound(nil))
			return
		}
		// TODO Check if exists
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// TODO Manage error properly
			log.Error(err)
			return
		}
		//defer func() { _ = conn.Close() }()
		go func() {
			ticker := time.NewTicker(pingPeriod)
			defer func() {
				ticker.Stop()
				_ = conn.Close()
			}()
			messages := make(chan *nats.Msg)
			sub, err := natsConn.ChanSubscribe(id, messages)
			if err != nil {
				// TODO Manage error properly
				log.Error(err)
				return
			}
			defer func() { _ = sub.Unsubscribe() }()
			for {
				select {
				case message, ok := <-messages:
					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if !ok {
						// Channel  closed
						_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}

					w, err := conn.NextWriter(websocket.TextMessage)
					if err != nil {
						// TODO Manage error properly
						log.Error(err)
						return
					}
					_, _ = w.Write(message.Data)

					if err := w.Close(); err != nil {
						// TODO Manage error properly
						log.Error(err)
						return
					}
				case <-ticker.C:
					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				}
			}
		}()

		// TODO send via service.
		m, _ := WriteEventJson(&PlayerJoint{EventBase: NewEventBase(table.Id, []string{"*"}, user), Player: user})
		_ = natsConn.Publish(id, m)
	}))
}
