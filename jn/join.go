package jn

type joins struct {
	t1   any
	join string
	t2   any
}

func (j joins) FirstArg() any {
	return j.t1
}

func (j joins) Join() string {
	return j.join
}

func (j joins) SecondArg() any {
	return j.t2
}

// Join makes a inner join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(jn.Join(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func Join[T any, U, V *T | **T](t1 U, t2 V) joins {
	return joins{t1: t1, join: "JOIN", t2: t2}
}

// LeftJoin makes a left join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(jn.LeftJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func LeftJoin[T any, U, V *T | **T](t1 U, t2 V) joins {
	return joins{t1: t1, join: "LEFT JOIN", t2: t2}
}

// RightJoin makes a right join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(jn.RightJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func RightJoin[T any, U, V *T | **T](t1 U, t2 V) joins {
	return joins{t1: t1, join: "RIGHT JOIN", t2: t2}
}

func CustomJoin(t1 any, j string, t2 any) joins {
	return joins{t1: t1, join: j, t2: t2}
}
