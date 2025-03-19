package where

import (
	"reflect"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
)

type valueOperation struct {
	value any
}

func (vo valueOperation) GetValue() any {
	if result, ok := vo.value.(model.ValueOperation); ok {
		return result.GetValue()
	}
	return vo.value
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
func Equals[T any, A *T | **T](a A, v T) model.Operation {
	if reflect.ValueOf(v).Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil() {
		return model.Operation{Arg: a, Operator: enum.Is, Type: enum.OperationIsWhere}
	}
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Equals, Type: enum.OperationWhere}
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
func NotEquals[T any, A *T | **T](a A, v T) model.Operation {
	if reflect.ValueOf(v).Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil() {
		return model.Operation{Arg: a, Operator: enum.IsNot, Type: enum.OperationIsWhere}
	}
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.NotEquals, Type: enum.OperationWhere}
}

// Greater creates a ">" to value inside a where clause
//
// # Example
//
//	// get all animals that was created after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Greater(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Greater[T any, A *T | **T](a A, v T) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Greater, Type: enum.OperationWhere}
}

// GreaterEquals creates a ">=" to value inside a where clause
//
// # Example
//
//	// get all animals that was created in or after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.GreaterEquals(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func GreaterEquals[T any, A *T | **T](a A, v T) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.GreaterEquals, Type: enum.OperationWhere}
}

// Less creates a "<" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Less(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Less[T any, A *T | **T](a A, v T) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Less, Type: enum.OperationWhere}
}

// LessEquals creates a "<=" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated in or before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.LessEquals(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func LessEquals[T any, A *T | **T](a A, v T) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.LessEquals, Type: enum.OperationWhere}
}

// Like creates a "LIKE" to value inside a where clause
//
// # Example
//
//	// get all animals that has a "at" in his name
//	db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%at%")).Scan(&a)
func Like[T any](a *T, v string) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Like, Type: enum.OperationWhere}
}

func In[T any, V []T | *model.Query](a *T, mq V) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: mq}, Operator: enum.In, Type: enum.OperationInWhere}
}

// And creates a "AND" inside a where clause
//
// # Example
//
//	// and can connect model.operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.And(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func And() model.Operation {
	return model.Operation{Operator: enum.And, Type: enum.LogicalWhere}
}

// Or creates a "OR" inside a where clause
//
// # Example
//
//	// or can connect model.operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.Or(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func Or() model.Operation {
	return model.Operation{Operator: enum.Or, Type: enum.LogicalWhere}
}

// # Example
//
//	Where(wh.EqualsArg(&db.Job.Id, &db.Person.Id))
func EqualsArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Equals, Type: enum.OperationAttributeWhere}
}

// # Example
//
//	Where(wh.NotEqualsArg(&db.Job.Id, &db.Person.Id))
func NotEqualsArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.NotEquals, Type: enum.OperationAttributeWhere}
}

// # Example
//
//	Where(wh.GreaterArg(&db.Stock.Minimum, &db.Drinks.Stock))
func GreaterArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Greater, Type: enum.OperationAttributeWhere}
}

// # Example
//
//	Where(wh.GreaterEqualsArg(&db.Drinks.Reorder, &db.Drinks.Stock))
func GreaterEqualsArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.GreaterEquals, Type: enum.OperationAttributeWhere}
}

// # Example
//
//	Where(wh.LessArg(&db.Exam.Score, &db.Data.Minimum))
func LessArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.Less, Type: enum.OperationAttributeWhere}
}

// # Example
//
//	Where(wh.LessEqualsArg(&db.Exam.Score, &db.Data.Minimum))
func LessEqualsArg(a any, v any) model.Operation {
	return model.Operation{Arg: a, Value: valueOperation{value: v}, Operator: enum.LessEquals, Type: enum.OperationAttributeWhere}
}
