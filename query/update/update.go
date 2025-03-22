package update

import "github.com/olauro/goe/model"

func Set[T any, A *T | **T](a A, v T) model.Set {
	return model.Set{Attribute: a, Value: v}
}
