package query

type Joins interface {
	FirstArg() any
	Join() string
	SecondArg() any
}

type join struct {
	t1   any
	join string
	t2   any
}

func (j join) FirstArg() any {
	return j.t1
}

func (j join) Join() string {
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
func Join[T any, U, V *T | **T](t1 U, t2 V) Joins {
	return join{t1: t1, join: "JOIN", t2: t2}
}

// LeftJoin makes a left join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(query.LeftJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func LeftJoin[T any, U, V *T | **T](t1 U, t2 V) Joins {
	return join{t1: t1, join: "LEFT JOIN", t2: t2}
}

// RightJoin makes a right join betwent the tables
//
// # Example
//
//	db.Select(db.Food).From(db.Food).
//	Joins(query.RightJoin(&db.Animal.Id, &db.Food.IdAnimal)).Scan(&a)
func RightJoin[T any, U, V *T | **T](t1 U, t2 V) Joins {
	return join{t1: t1, join: "RIGHT JOIN", t2: t2}
}

func CustomJoin(t1 any, j string, t2 any) Joins {
	return join{t1: t1, join: j, t2: t2}
}
