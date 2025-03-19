package function

import (
	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/query"
)

// ToUpper uses database function to converts the target string to uppercase
//
// # Example
//
//	goe.Select(&struct {
//		UpperName *query.Function[string]
//	}{
//		UpperName: function.ToUpper(&db.Animal.Name),
//	}).From(db.Animal)
func ToUpper(target *string) *query.Function[string] {
	return &query.Function[string]{Field: target, Type: enum.UpperFunction}
}

// Argument is used to pass a value to a function inside a where clause
//
// # Example
//
//	goe.Select(db.Animal).From(db.Animal).Wheres(where.Equals(function.ToUpper(&db.Animal.Name), function.Argument("CAT"))).AsSlice()
func Argument[T any](value T) query.Function[T] {
	return query.Function[T]{Value: value}
}
