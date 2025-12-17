package function

import (
	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

// ToUpper uses database function to converts the target string to uppercase
//
// # Example
//
//	goe.Select(&struct {
//		UpperName *function[string]
//	}{
//		UpperName: function.ToUpper(&db.Animal.Name),
//	})
func ToUpper(target *string) *function[string] {
	return &function[string]{Field: target, Type: enum.UpperFunction}
}

// ToLower uses database function to converts the target string to lowercase
//
// # Example
//
//	goe.Select(&struct {
//		LowerName *function[string]
//	}{
//		LowerName: function.ToLower(&db.Animal.Name),
//	})
func ToLower(target *string) *function[string] {
	return &function[string]{Field: target, Type: enum.LowerFunction}
}

// Argument is used to pass a value to a function inside a where clause
//
// # Example
//
//	goe.Select(db.Animal).Where(where.Equals(function.ToUpper(&db.Animal.Name), function.Argument("CAT"))).AsSlice()
func Argument[T any](value T) function[T] {
	return function[T]{Value: value}
}

type function[T any] struct {
	Field *T
	Type  enum.FunctionType
	Value T
}

func (f function[T]) GetValue() any {
	return f.Value
}

func (f function[T]) GetType() enum.FunctionType {
	return f.Type
}

func (f function[T]) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:        b.Table,
		Name:         b.Name,
		FunctionType: f.Type,
	}
}

func (f function[T]) GetField() any {
	return f.Field
}
