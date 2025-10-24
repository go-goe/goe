package standard

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/repository"
	"github.com/go-goe/goe"
)

type handler struct {
	repository repository.Repository[data.Person]
}

type Error struct {
	Status      int    `json:"status"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

func NewPersonHandler(r repository.Repository[data.Person]) handler {
	return handler{r}
}

func Use(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := f(w, r); err != nil {
			log.Printf("error %v on %v", err, r.RequestURI)
		}
	}
}

func (h handler) Find(w http.ResponseWriter, r *http.Request) error {
	var err error
	var id int
	if id, err = strconv.Atoi(r.PathValue("id")); err != nil {
		return notFoundError(w, err)
	}

	var p *data.Person
	if p, err = h.repository.Find(data.Person{ID: id}); err != nil {
		if errors.Is(err, goe.ErrNotFound) {
			return notFoundError(w, err)
		}
		return internalServerError(w, err)
	}

	if err = json.NewEncoder(w).Encode(p); err != nil {
		return internalServerError(w, err)
	}
	return nil
}

func (h handler) List(w http.ResponseWriter, r *http.Request) error {
	var err error
	var page, size int
	w.Header().Set("Content-Type", "application/json")
	if page, err = strconv.Atoi(r.FormValue("page")); err != nil {
		return badRequestError(w, err, "invalid page")
	}

	if size, err = strconv.Atoi(r.FormValue("size")); err != nil {
		fmt.Println(page, size)
		return badRequestError(w, err, "invalid size")
	}

	var pages *goe.Pagination[data.Person]
	if pages, err = h.repository.List(page, size); err != nil {
		return internalServerError(w, err)
	}

	if err = json.NewEncoder(w).Encode(pages); err != nil {
		return internalServerError(w, err)
	}
	return nil
}

func (h handler) Create(w http.ResponseWriter, r *http.Request) error {
	var p data.Person
	var err error
	if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
		return internalServerError(w, err)

	}

	if err = h.repository.Create(&p); err != nil {
		return internalServerError(w, err)
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(struct {
		ID int `json:"id"`
	}{ID: p.ID}); err != nil {
		return internalServerError(w, err)
	}
	return nil
}

func (h handler) Save(w http.ResponseWriter, r *http.Request) error {
	var p data.Person
	var err error

	if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
		return internalServerError(w, err)
	}

	if p.ID, err = strconv.Atoi(r.PathValue("id")); err != nil {
		return badRequestError(w, err, "invalid id")
	}

	if err = h.repository.Save(p); err != nil {
		return internalServerError(w, err)
	}

	if err = ok(w); err != nil {
		return internalServerError(w, err)
	}
	return nil
}

func (h handler) Remove(w http.ResponseWriter, r *http.Request) error {
	var p data.Person
	var err error

	if p.ID, err = strconv.Atoi(r.PathValue("id")); err != nil {
		return badRequestError(w, err, "invalid id")
	}

	if err = h.repository.Remove(p); err != nil {
		return internalServerError(w, err)
	}

	if err = ok(w); err != nil {
		return internalServerError(w, err)
	}
	return nil
}

func ok(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(Error{Status: http.StatusOK, Message: "200 Ok"})
}

func badRequestError(w http.ResponseWriter, err error, des string) error {
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(Error{Status: http.StatusBadRequest, Message: "400 Bad Request", Description: des}); err != nil {
		return err
	}
	return err
}

func notFoundError(w http.ResponseWriter, err error) error {
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(Error{Status: http.StatusNotFound, Message: "404 Not Found"}); err != nil {
		return err
	}
	return err
}

func internalServerError(w http.ResponseWriter, err error) error {
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(Error{Status: http.StatusNotFound, Message: "500 Internal Server Error"}); err != nil {
		return err
	}
	return err
}
