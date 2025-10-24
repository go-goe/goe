package data

import (
	"github.com/go-goe/goe"
	"github.com/go-goe/sqlite"
)

type Person struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Database struct {
	Person *Person
	*goe.DB
}

func NewDatabase() (*Database, error) {
	db, err := goe.Open[Database](sqlite.Open("crud-basic.db", sqlite.Config{}))
	if err != nil {
		return nil, err
	}

	err = goe.Migrate(db).AutoMigrate()
	if err != nil {
		return nil, err
	}
	return db, nil
}
