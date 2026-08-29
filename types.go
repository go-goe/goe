package goe

import (
	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/query/where"
)

type customWhere struct {
	modelWhere model.Where
}

func (cw customWhere) And(c customWhere) model.Where {
	return where.And(cw.modelWhere, c.modelWhere)
}

type Type[T any] struct {
	field
	value T
}

func (t Type[T]) getField() field {
	return t.field
}

type TypeNull[T any] struct {
	field
	value *T
}

func (t TypeNull[T]) getField() field {
	return t.field
}

func (t Type[T]) Equals(v T) {
}

func (t Type[T]) Join(v TypeInterface[T]) (field, field) {
	return t.field, v.getField()
}

type ManyToOne[Many any, One any] struct {
	field
	many Many
	one  One
}
