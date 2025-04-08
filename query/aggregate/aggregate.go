package aggregate

import "github.com/go-goe/goe/query"

// Aggregate Count uses database aggregate to make a count on the target
//
// Is used with [query.Count] as argument for select
//
// # Example
//
//	goe.Select(&struct {
//		*query.Count
//	}{
//		aggregate.Count(&db.Animal.Id),
//	}).From(db.Animal)
func Count(t any) *query.Count {
	return &query.Count{Field: t}
}
