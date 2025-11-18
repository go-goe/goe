package goe

import (
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
