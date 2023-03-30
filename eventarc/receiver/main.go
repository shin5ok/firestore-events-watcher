package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/go-chi/render"
)

var appName = "myapp"

var servicePort = os.Getenv("PORT")

func main() {

	oplog := httplog.LogEntry(context.Background())
	/* jsonify logging */
	httpLogger := httplog.NewLogger(appName, httplog.Options{JSON: true, LevelFieldName: "severity", Concise: true})

	/* exporter for prometheus */
	m := chiprometheus.NewMiddleware(appName)

	r := chi.NewRouter()
	// r.Use(middleware.Throttle(8))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httplog.RequestLogger(httpLogger))
	r.Use(m)

	r.Post("/pub", func(w http.ResponseWriter, r *http.Request) {
		subject := r.Header.Get("Ce-Subject")
		if subject == "" {
			errorRender(w, r, 500, fmt.Errorf("cannot get Ce-Subject"))
		}
		log.Printf("Ce-Subject: %+v\n", subject)
		render.JSON(w, r, map[string]any{"Ce-Subject": subject})

	})

	r.Post("/pub-detail", func(w http.ResponseWriter, r *http.Request) {
		allHeaders := r.Header
		log.Printf("Ce-Subject: %+v\n", allHeaders)
		render.JSON(w, r, allHeaders)

	})

	if err := http.ListenAndServe(":"+servicePort, r); err != nil {
		oplog.Err(err)
	}

}

var errorRender = func(w http.ResponseWriter, r *http.Request, httpCode int, err error) {
	render.Status(r, httpCode)
	render.JSON(w, r, map[string]interface{}{"ERROR": err.Error()})
}
