package function

import (
	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/query"
)

func ToUpper(target *string) *query.Function[string] {
	return &query.Function[string]{Field: target, Type: enum.UpperFunction}
}

func Argument[T any](value T) query.Function[T] {
	return query.Function[T]{Value: value}
}
