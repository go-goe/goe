package goe

import (
	"github.com/go-goe/goe/model"
)

type Type[T any] struct {
	field
	value   T
	builder *builder
}

func (t Type[T]) getField() field {
	return t.field
}

func (t Type[T]) Equals(v T) customWhere {
	return equalsWhere(v, t.field)
}

func (t Type[T]) Like(v string) customWhere {
	return likeWhere(v, t.field)
}

func (t Type[T]) Set(v T) model.Set {
	return model.Set{Attribute: t.field, Value: v}
}

func (t Type[T]) Join(v TypeInterface[T]) (field, field) {
	return t.field, v.getField()
}

func (t Type[T]) Value() T {
	if t.builder != nil {
		t.builder.fieldsSelect = append(t.builder.fieldsSelect, t.field)
	}
	return t.value
}

func (t *Type[T]) setBuilder(builder *builder) {
	t.builder = builder
}

type TypeNull[T any] struct {
	field
	value   *T
	builder *builder
}

func (t TypeNull[T]) getField() field {
	return t.field
}

func (t TypeNull[T]) Equals(v *T) customWhere {
	return equalsNilWhere(v, t.field)
}

func (t TypeNull[T]) Set(v *T) model.Set {
	return model.Set{Attribute: t.field, Value: v}
}

func (t TypeNull[T]) Join(v TypeInterface[T]) (field, field) {
	return t.field, v.getField()
}

func (t TypeNull[T]) Value() *T {
	if t.builder != nil {
		t.builder.fieldsSelect = append(t.builder.fieldsSelect, t.field)
	}
	return t.value
}

func (t *TypeNull[T]) setBuilder(builder *builder) {
	t.builder = builder
}
