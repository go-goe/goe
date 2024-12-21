package wh

import (
	"fmt"
	"reflect"
)

type operator interface {
	Operation() string
}

type Operation struct {
	Arg       any
	Value     any
	Operator  string
	ValueFlag string
}

func (o Operation) Operation() string {
	return fmt.Sprintf("%v %v %v", o.Arg, o.Operator, o.ValueFlag)
}

type OperationIs struct {
	Arg      any
	Operator string
}

func (o OperationIs) Operation() string {
	return fmt.Sprintf("%v %v NULL", o.Arg, o.Operator)
}

// Equals creates a "=" to v inside a where clause or makes a IS when v is nil
//
// # Example
//
//	// delete Food with Id fc1865b4-6f2d-4cc6-b766-49c2634bf5c4
//	db.Delete(db.Food).Where(wh.Equals(&db.Food.Id, "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4"))
//
//	// implicit join using where equals
//	db.Select(db.Animal).
//	From(db.Animal, db.AnimalFood, db.Food).
//	Where(
//		wh.Equals(&db.Animal.Id, &db.AnimalFood.IdAnimal),
//		wh.And(),
//		wh.Equals(&db.Food.Id, &db.AnimalFood.IdFood)).
//	Scan(&a)
//
//	// generate: WHERE "animals"."idhabitat" IS NULL
//	Where(wh.Equals(&db.Animal.IdHabitat, nil)).Scan(&a)
func Equals[T any, A *T | **T](a A, v T) operator {
	if reflect.ValueOf(v).Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil() {
		return OperationIs{Arg: a, Operator: "IS"}
	}
	return Operation{Arg: a, Value: v, Operator: "="}
}

// NotEquals creates a "<>" to value inside a where clause
//
// # Example
//
//	// get all foods that name are not Cookie
//	db.Select(db.Food).From(db.Animal).
//	Where(wh.NotEquals(&db.Food.Name, "Cookie")).Scan(&f)
//
//	// generate: WHERE "animals"."idhabitat" IS NOT NULL
//	Where(wh.NotEquals(&db.Animal.IdHabitat, nil)).Scan(&a)
func NotEquals[T any, A *T | **T](a A, v T) operator {
	if reflect.ValueOf(v).Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil() {
		return OperationIs{Arg: a, Operator: "IS NOT"}
	}
	return Operation{Arg: a, Value: v, Operator: "<>"}
}

// Greater creates a ">" to value inside a where clause
//
// # Example
//
//	// get all animals that was created after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Greater(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Greater[T any, A *T | **T](a A, v T) Operation {
	return Operation{Arg: a, Value: v, Operator: ">"}
}

// GreaterEquals creates a ">=" to value inside a where clause
//
// # Example
//
//	// get all animals that was created in or after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.GreaterEquals(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func GreaterEquals[T any, A *T | **T](a A, v T) Operation {
	return Operation{Arg: a, Value: v, Operator: ">="}
}

// Less creates a "<" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Less(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Less[T any, A *T | **T](a A, v T) Operation {
	return Operation{Arg: a, Value: v, Operator: "<"}
}

// LessEquals creates a "<=" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated in or before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.LessEquals(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func LessEquals[T any, A *T | **T](a A, v T) Operation {
	return Operation{Arg: a, Value: v, Operator: "<="}
}

// Like creates a "LIKE" to value inside a where clause
//
// # Example
//
//	// get all animals that has a "at" in his name
//	db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%at%")).Scan(&a)
func Like[T any](a *T, v string) Operation {
	return Operation{Arg: a, Value: v, Operator: "LIKE"}
}

// Not creates a "NOT" inside a where clause
//
// # Example
//
//	// get all animals that not has a "at" in his name
//	db.Select(db.Animal).From(db.Animal).Where(wh.Not(wh.Like(&db.Animal.Name, "%at%"))).Scan(&a)
func Not(o Operation) Operation {
	o.Operator = "NOT " + o.Operator
	return o
}

func NewOperator[T any](a *T, operator string, v T) Operation {
	return Operation{Arg: a, Value: v, Operator: operator}
}

type Logical struct {
	Operator string
}

func (l Logical) Operation() string {
	return fmt.Sprintf(" %v ", l.Operator)
}

// And creates a "AND" inside a where clause
//
// # Example
//
//	// and can connect operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.And(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func And() Logical {
	return Logical{Operator: "AND"}
}

// Or creates a "OR" inside a where clause
//
// # Example
//
//	// or can connect operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.Or(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func Or() Logical {
	return Logical{Operator: "OR"}
}

type OperationArg struct {
	Op Operation
}

func (o OperationArg) Operation() string {
	return fmt.Sprintf("%v %v %v", o.Op.Arg, o.Op.Operator, o.Op.ValueFlag)
}

// # Example
//
//	Where(wh.EqualsArg(&db.Job.Id, &db.Person.Id))
func EqualsArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: "="}}
}

// # Example
//
//	Where(wh.NotEqualsArg(&db.Job.Id, &db.Person.Id))
func NotEqualsArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: "<>"}}
}

// # Example
//
//	Where(wh.GreaterArg(&db.Stock.Minimum, &db.Drinks.Stock))
func GreaterArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: ">"}}
}

// # Example
//
//	Where(wh.GreaterEqualsArg(&db.Drinks.Reorder, &db.Drinks.Stock))
func GreaterEqualsArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: ">="}}
}

// # Example
//
//	Where(wh.LessArg(&db.Exam.Score, &db.Data.Minimum))
func LessArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: "<"}}
}

// # Example
//
//	Where(wh.LessEqualsArg(&db.Exam.Score, &db.Data.Minimum))
func LessEqualsArg(a any, v any) OperationArg {
	return OperationArg{Operation{Arg: a, Value: v, Operator: "<="}}
}
