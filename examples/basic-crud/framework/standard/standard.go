package standard

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/framework"
	"github.com/go-goe/examples/crud-basic/handler/standard"
	"github.com/go-goe/examples/crud-basic/repository"
)

type standardStarter struct {
	db *data.Database
}

func NewStarter(db *data.Database) framework.Starter {
	return standardStarter{db}
}

func LoggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		next.ServeHTTP(w, r)

		logger.Info("request info", "method", r.Method, "path", r.URL.Path, "duration", time.Since(startTime).String())
	})
}

func (s standardStarter) Start(port string) error {
	mux, err := s.Route()
	if err != nil {
		return err
	}

	log.Println("server running on", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		return err
	}
	return nil
}

func (s standardStarter) Route() (http.Handler, error) {
	personHandler := standard.NewHandler(repository.NewRepository(s.db.Person))
	var logger *slog.Logger

	mux := http.NewServeMux()

	mux.HandleFunc("GET /persons/{id}", standard.Use(personHandler.Find, logger))
	mux.HandleFunc("POST /persons", standard.Use(personHandler.Create, logger))
	mux.Handle("GET /persons", standard.Use(personHandler.List, logger))
	mux.HandleFunc("PUT /persons/{id}", standard.Use(personHandler.Save, logger))
	mux.HandleFunc("DELETE /persons/{id}", standard.Use(personHandler.Remove, logger))

	if os.Getenv("GO_ENV") != "test" {
		logger = framework.NewLogger()
		return LoggingMiddleware(mux, logger), nil
	}

	return mux, nil
}
