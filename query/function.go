package query

import (
	"fmt"

	"github.com/olauro/goe/enum"
)

type Function[T any] struct {
	Field *T
	Type  enum.FunctionType
	Value T
}

func (f *Function[string]) Scan(src any) error {
	v, ok := src.(string)
	if !ok {
		return fmt.Errorf("error scan function")
	}

	f.Value = v
	return nil
}

func ToUpper(target *string) *Function[string] {
	return &Function[string]{Field: target, Type: enum.UpperFunction}
}

func Argument[T any](value T) Function[T] {
	return Function[T]{Value: value}
}

func (f Function[T]) GetValue() any {
	return f.Value
}
