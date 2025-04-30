package query

import (
	"fmt"
	"reflect"

	"github.com/go-goe/goe/enum"
)

type Function[T any] struct {
	Field *T
	Type  enum.FunctionType
	Value T
}

func (f *Function[T]) Scan(src any) error {
	v, ok := src.(T)
	if !ok {
		return fmt.Errorf("error scan function")
	}

	f.Value = v
	return nil
}

func (f Function[T]) GetValue() any {
	return f.Value
}

func (f Function[T]) GetType() enum.FunctionType {
	return f.Type
}

type Count struct {
	Field any
	Value int64
}

func (c *Count) Scan(src any) error {
	v, ok := src.(int64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	c.Value = v
	return nil
}

func (c Count) Aggregate() enum.AggregateType {
	return enum.CountAggregate
}

// Get removes the pointer used when [goe.Select] get any argument.
//
// # Example
//
//	for row, err := range goe.Select(&struct {
//		User    *string //goe needs a pointer for store the referecent argument
//		Role    *string //goe needs a pointer for store the referecent argument
//		EndTime **time.Time //if the argument is already a pointer goe needs a pointer to a pointer for store the referecent argument
//	}{
//		User:    &db.User.Name,
//		Role:    &db.Role.Name,
//		EndTime: &db.UserRole.EndDate,
//	}).
//	Joins(
//		join.LeftJoin[int](&db.User.Id, &db.UserRole.UserId),
//		join.LeftJoin[int](&db.UserRole.RoleId, &db.Role.Id),
//	).
//	OrderByAsc(&db.User.Id).Rows() {
//		q = append(q, struct {
//			User    string //return model can be different from select model
//			Role    string
//			EndTime *time.Time
//		}{
//			User:    query.Get(row.User), //get a empty string if the database returns null
//			Role:    query.Get(row.Role), //get a empty string if the database returns null
//			EndTime: query.Get(row.EndTime), //EndTime can store nil/null values
//		})
//	}
func Get[T any](v *T) T {
	if v == nil {
		return reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface().(T)
	}
	return *v
}
