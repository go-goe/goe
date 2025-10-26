package gin

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/framework"
	ginHandler "github.com/go-goe/examples/crud-basic/handler/gin"
	"github.com/go-goe/examples/crud-basic/repository"
)

type ginStarter struct {
	db *data.Database
}

func NewStarter(db *data.Database) framework.Starter {
	return ginStarter{db}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		logger.Info("request info", "method", c.Request.Method, "path", c.Request.URL.Path, "clientIP", c.ClientIP(), "duration", time.Since(startTime).String())
	}
}

func (s ginStarter) Start(port string) error {
	r, err := s.Route()
	if err != nil {
		return err
	}

	rg := r.(*gin.Engine)

	if err = rg.Run(":" + port); err != nil {
		return err
	}
	return nil
}

func (s ginStarter) Route() (http.Handler, error) {
	personHandler := ginHandler.NewHandler(repository.NewRepository(s.db.Person))

	var r *gin.Engine
	var logger *slog.Logger
	if os.Getenv("GO_ENV") == "test" {
		gin.SetMode(gin.TestMode)
		r = gin.New()
	} else {
		logger = framework.NewLogger()
		r = gin.New()
		r.Use(LoggingMiddleware(logger))
	}

	r.GET("/persons/:id", ginHandler.Use(personHandler.Find, logger))
	r.POST("/persons", ginHandler.Use(personHandler.Create, logger))
	r.GET("/persons", ginHandler.Use(personHandler.List, logger))
	r.PUT("/persons/:id", ginHandler.Use(personHandler.Save, logger))
	r.DELETE("/persons/:id", ginHandler.Use(personHandler.Remove, logger))

	return r, nil
}
