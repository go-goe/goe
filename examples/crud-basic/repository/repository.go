package repository

import (
	"github.com/go-goe/goe"
)

type Repository[E any] interface {
	Find(t E) (*E, error)
	Create(t *E) error
	List(page, size int) (*goe.Pagination[E], error)
	Save(t E) error
	Remove(t E) error
}

type repository[E any] struct {
	entity *E
}

func NewRepository[E any](entity *E) Repository[E] {
	return repository[E]{entity: entity}
}

func (r repository[E]) Find(t E) (*E, error) {
	return goe.Find(r.entity).ById(t)
}

func (r repository[E]) Create(t *E) error {
	return goe.Insert(r.entity).One(t)
}

func (r repository[E]) List(page, size int) (*goe.Pagination[E], error) {
	return goe.List(r.entity).AsPagination(page, size)
}

func (r repository[E]) Save(t E) error {
	return goe.Save(r.entity).ByValue(t)
}

func (r repository[E]) Remove(t E) error {
	return goe.Remove(r.entity).ById(t)
}
