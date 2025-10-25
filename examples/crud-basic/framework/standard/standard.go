package standard

import (
	"log"
	"net/http"
	"os"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/handler/standard"
	"github.com/go-goe/examples/crud-basic/repository"
)

func Start(db *data.Database) error {
	mux, err := Router(db)
	if err != nil {
		return err
	}

	port := os.Getenv("PORT")
	log.Println("server running on", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		return err
	}
	return nil
}

func Router(db *data.Database) (http.Handler, error) {
	personHandler := standard.NewHandler(repository.NewRepository(db.Person))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /persons/{id}", standard.Use(personHandler.Find))
	mux.HandleFunc("POST /persons", standard.Use(personHandler.Create))
	mux.Handle("GET /persons", standard.Use(personHandler.List))
	mux.HandleFunc("PUT /persons/{id}", standard.Use(personHandler.Save))
	mux.HandleFunc("DELETE /persons/{id}", standard.Use(personHandler.Remove))

	return mux, nil
}
