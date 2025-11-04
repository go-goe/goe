package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
	"github.com/go-goe/goe"
	"github.com/go-goe/sqlite"
)

type Animal struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

type Database struct {
	Animal *Animal
	*goe.DB
}

type RequestAnimal struct {
	Name  string `json:"name" validate:"required"`
	Emoji string `json:"emoji" validate:"required"`
}

func main() {

	var db *Database
	var err error
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if db, err = goe.Open[Database](sqlite.Open("one-file-crud.db", sqlite.Config{DatabaseConfig: goe.DatabaseConfig{
		Logger: logger,
	}})); err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(logger)

	goe.Migrate(db).AutoMigrate()

	s := fuego.NewServer()

	fuego.Get(s, "/animals/{id}", func(c fuego.ContextNoBody) (*Animal, error) {
		return goe.Find(db.Animal).ByID(Animal{ID: c.PathParamInt("id")})
	}, option.Path("id", "animal id", param.Integer()))

	fuego.Get(s, "/animals", func(c fuego.ContextNoBody) (*goe.Pagination[Animal], error) {
		return goe.List(db.Animal).OrderByAsc(&db.Animal.ID).Match(Animal{Name: c.QueryParam("name")}).
			AsPagination(c.QueryParamInt("page"), c.QueryParamInt("size"))
	}, option.QueryInt("page", "current page"), option.QueryInt("size", "page size"), option.Query("name", "animal name"))

	fuego.Post(s, "/animals", func(c fuego.ContextWithBody[RequestAnimal]) (any, error) {
		request, err := c.Body()
		if err != nil {
			return nil, fuego.BadRequestError{}
		}
		return nil, goe.Insert(db.Animal).One(&Animal{Name: request.Name, Emoji: request.Emoji})
	})

	fuego.Put(s, "/animals/{id}", func(c fuego.ContextWithBody[RequestAnimal]) (any, error) {
		request, err := c.Body()
		if err != nil {
			return nil, fuego.BadRequestError{}
		}
		return nil, goe.Save(db.Animal).ByID(Animal{ID: c.PathParamInt("id"), Name: request.Name, Emoji: request.Emoji})
	}, option.Path("id", "animal id", param.Integer()))

	fuego.Delete(s, "/animals/{id}", func(c fuego.ContextNoBody) (any, error) {
		return nil, goe.Remove(db.Animal).ByID(Animal{ID: c.PathParamInt("id")})
	}, option.Path("id", "animal id", param.Integer()))

	s.Run()
}
