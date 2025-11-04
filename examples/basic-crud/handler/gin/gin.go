package gin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/handler"
	"github.com/go-goe/examples/crud-basic/repository"
	"github.com/go-goe/goe"
)

type handlerGin struct {
	repository repository.Repository[data.Person]
}

func NewHandler(r repository.Repository[data.Person]) handlerGin {
	return handlerGin{r}
}

func Use(f func(c *gin.Context) error, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil && logger != nil {
			logger.Error("request error", "error", err, "uri", c.Request.RequestURI)
		}
	}
}

func (h handlerGin) Find(c *gin.Context) error {
	var err error
	var p *data.Person
	var id int
	if id, err = strconv.Atoi(c.Param("id")); err != nil {
		badRequestError(c, "invalid id")
		return err
	}

	if p, err = h.repository.Find(data.Person{ID: id}); err != nil {
		if errors.Is(err, goe.ErrNotFound) {
			notFoundError(c)
			return err
		}
		internalServerError(c)
		return err
	}

	ok(c, p)
	return nil
}

func (h handlerGin) List(c *gin.Context) error {
	var err error
	var pages *goe.Pagination[data.Person]
	var page, size int

	if page, err = strconv.Atoi(c.Query("page")); err != nil {
		badRequestError(c, "invalid page")
		return err
	}

	if size, err = strconv.Atoi(c.Query("size")); err != nil {
		badRequestError(c, "invalid size")
		return err
	}

	if pages, err = h.repository.List(page, size); err != nil {
		internalServerError(c)
		return err
	}

	ok(c, pages)
	return nil
}

func (h handlerGin) Create(c *gin.Context) error {
	var p data.Person
	var err error
	if err = json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
		internalServerError(c)
		return err
	}

	if err = h.repository.Create(&p); err != nil {
		internalServerError(c)
		return err
	}

	created(c, p.ID)
	return nil
}

func (h handlerGin) Save(c *gin.Context) error {
	var p data.Person
	var err error
	var id int

	if id, err = strconv.Atoi(c.Param("id")); err != nil {
		badRequestError(c, "invalid id")
		return err
	}

	if err = json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
		internalServerError(c)
		return err
	}
	p.ID = id

	if err = h.repository.Save(p); err != nil {
		internalServerError(c)
		return err
	}

	ok(c, nil)
	return nil
}

func (h handlerGin) Remove(c *gin.Context) error {
	var err error
	var id int

	if id, err = strconv.Atoi(c.Param("id")); err != nil {
		badRequestError(c, "invalid id")
		return err
	}

	if err := h.repository.Remove(data.Person{ID: id}); err != nil {
		internalServerError(c)
		return err
	}

	ok(c, nil)
	return nil
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, handler.Response[any]{Status: http.StatusOK, Message: "200 Ok", Data: data})
}

func created[T any](c *gin.Context, id T) {
	c.JSON(http.StatusCreated, handler.Response[handler.ResponsePost[T]]{Status: http.StatusOK, Message: "201 Created", Data: handler.ResponsePost[T]{ID: id}})
}

func badRequestError(c *gin.Context, des string) {
	c.JSON(http.StatusBadRequest, handler.Response[any]{Status: http.StatusBadRequest, Message: "400 Bad Request", Description: des})
}

func notFoundError(c *gin.Context) {
	c.JSON(http.StatusNotFound, handler.Response[any]{Status: http.StatusNotFound, Message: "404 Not Found"})
}

func internalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, handler.Response[any]{Status: http.StatusInternalServerError, Message: "500 Internal Server Error"})
}
