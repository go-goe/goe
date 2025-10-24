package standard

import (
	"log"
	"net/http"
	"os"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/handler/standard"
	"github.com/go-goe/examples/crud-basic/repository"
)

func Start() error {
	db, err := data.NewDatabase()
	if err != nil {
		return err
	}
	personHandler := standard.NewPersonHandler(repository.NewRepository(db.Person))

	http.HandleFunc("GET /persons/{id}", standard.Use(personHandler.Find))
	http.HandleFunc("POST /persons", standard.Use(personHandler.Create))
	http.Handle("GET /persons", standard.Use(personHandler.List))
	http.HandleFunc("PUT /persons/{id}", standard.Use(personHandler.Save))
	http.HandleFunc("DELETE /persons/{id}", standard.Use(personHandler.Remove))

	port := os.Getenv("PORT")
	log.Println("server running on", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}
	return nil
}
