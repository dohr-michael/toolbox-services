package admin

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/rpg-tools/toolbox-services/app_context"
	"net/http"
)

func Router(router chi.Router) {
	router.Post("/push-to-pubsub", func(w http.ResponseWriter, r *http.Request) {
		nc := app_context.GetNats(r.Context())
		err := nc.Publish("admin", []byte("message"))
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
}
