package join

import (
	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type join struct {
	t1   any
	join enum.JoinType
	t2   any
}

func (j join) FirstArg() any {
	return j.t1
}

func (j join) Join() enum.JoinType {
	return j.join
}

func (j join) SecondArg() any {
	return j.t2
}

// Join makes a inner join betwent the tables
//
// # Example
//
//	goe.Select(db.Food).
//	Joins(join.Join(&db.Animal.Id, &db.Food.IdAnimal)).AsSlice()
func Join[T any, U, V *T | **T](t1 U, t2 V) model.Joins {
	return join{t1: t1, join: enum.Join, t2: t2}
}

// LeftJoin makes a left join betwent the tables
//
// # Example
//
//	goe.Select(db.Food).
//	Joins(join.LeftJoin(&db.Animal.Id, &db.Food.IdAnimal)).AsSlice()
func LeftJoin[T any, U, V *T | **T](t1 U, t2 V) model.Joins {
	return join{t1: t1, join: enum.LeftJoin, t2: t2}
}

// RightJoin makes a right join betwent the tables
//
// # Example
//
//	goe.Select(db.Food).
//	Joins(join.RightJoin(&db.Animal.Id, &db.Food.IdAnimal)).AsSlice()
func RightJoin[T any, U, V *T | **T](t1 U, t2 V) model.Joins {
	return join{t1: t1, join: enum.RightJoin, t2: t2}
}
