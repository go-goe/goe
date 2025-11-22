package goe

import (
	"context"

	"github.com/go-goe/goe/model"
)

type field interface {
	fieldSelect
	fieldDb
	isPrimaryKey() bool
	getTableId() int
	getFieldId() int
	getDefault() bool
	getAttributeName() string
	buildAttributeInsert(*builder)
	getSchemaID() int
	getEntityID() int
}

type fieldDb interface {
	getDb() *DB
}

type fieldSelect interface {
	fieldDb
	buildAttributeSelect([]model.Attribute, int)
	table() string
	schema() *string
	getTableId() int
}

type wherer[T any] interface {
	Equals(T) wherer[T]
	finalizer[T]
}

type finalizer[T any] interface {
	Delete() error
	DeleteContext(context.Context) error
	List() ([]T, error)
	ListContext(context.Context) ([]T, error)
	Update(T, ...any) error
	UpdateContext(context.Context, T, ...any) error
}
