package goe

import "github.com/go-goe/goe/model"

type Type[T any] struct {
	field
	value T
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

type TypeNull[T any] struct {
	field
	value *T
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
