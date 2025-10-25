package main

import (
	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/framework/standard"
	"github.com/go-goe/goe"
)

func main() {
	db, err := data.NewDatabase("crud-basic_test.db")
	if err != nil {
		panic(err)
	}
	defer goe.Close(db)
	standard.Start(db)
}
