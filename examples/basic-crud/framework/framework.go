package framework

import (
	"log/slog"
	"net/http"
	"os"
)

type Starter interface {
	Start(port string) error
	Router
}

type Router interface {
	Route() (http.Handler, error)
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
