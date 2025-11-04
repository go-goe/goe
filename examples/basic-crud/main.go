package main

import (
	"os"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/framework"
	ginFramework "github.com/go-goe/examples/crud-basic/framework/gin"
	"github.com/go-goe/examples/crud-basic/framework/standard"
	"github.com/go-goe/goe"
)

var frameworks map[string]func(db *data.Database) framework.Starter = map[string]func(db *data.Database) framework.Starter{
	"standard": standard.NewStarter,
	"gin":      ginFramework.NewStarter,
}

func main() {
	db, err := data.NewDatabase("crud-basic.db")
	if err != nil {
		panic(err)
	}
	defer goe.Close(db)

	starter := frameworks[os.Getenv("PK")]
	if starter == nil {
		panic("invalid package")
	}
	starter(db).Start(os.Getenv("PORT"))
}
