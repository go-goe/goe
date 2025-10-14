package aggregate

import "github.com/go-goe/goe/query"

// Aggregate Count uses database aggregate to make a count on the target
//
// Is used with [query.Count] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		query.Count
//	}](&struct {
//		*query.Count
//	}{
//		aggregate.Count(&db.Animal.Id),
//	})
func Count(t any) *query.Count {
	return &query.Count{Field: t}
}

func Avg(t any) *query.Avg {
	return &query.Avg{Field: t}
}

func Max(t any) *query.Max {
	return &query.Max{Field: t}
}

func Min(t any) *query.Min {
	return &query.Min{Field: t}
}

func Sum(t any) *query.Sum {
	return &query.Sum{Field: t}
}
