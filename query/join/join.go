package join

import (
	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/query"
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
//	db.Select(db.Food).From(db.Food).
//	Joins(query.Join(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func Join[T any, U, V *T | **T](t1 U, t2 V) query.Joins {
	return join{t1: t1, join: enum.Join, t2: t2}
}

// LeftJoin makes a left join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(query.LeftJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func LeftJoin[T any, U, V *T | **T](t1 U, t2 V) query.Joins {
	return join{t1: t1, join: enum.LeftJoin, t2: t2}
}

// RightJoin makes a right join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(query.RightJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func RightJoin[T any, U, V *T | **T](t1 U, t2 V) query.Joins {
	return join{t1: t1, join: enum.RightJoin, t2: t2}
}
