package aggregate

import "github.com/go-goe/goe/query"

// Aggregate Count uses database aggregate to make a count on the target
//
// Is used with [query.Count] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		Count query.Count
//	}](&struct {
//		Count *query.Count
//	}{
//		Count: aggregate.Count(&db.Animal.Id),
//	})
func Count(t any) *query.Count {
	return &query.Count{Field: t}
}

// Aggregate Avg uses database aggregate to get a average on the target
//
// Is used with [query.Avg] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		Avg query.Avg
//	}](&struct {
//		Avg *query.Avg
//	}{
//		Avg: aggregate.Avg(&db.Animal.Id),
//	})
func Avg(t any) *query.Avg {
	return &query.Avg{Field: t}
}

// Aggregate Max uses database aggregate to get the maximum value on the target
//
// Is used with [query.Max] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		Max query.Max
//	}](&struct {
//		Max *query.Max
//	}{
//		Max: aggregate.Max(&db.Animal.Id),
//	})
func Max(t any) *query.Max {
	return &query.Max{Field: t}
}

// Aggregate Min uses database aggregate to get the minimum value on the target
//
// Is used with [query.Min] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		Min query.Min
//	}](&struct {
//		Min *query.Min
//	}{
//		Min: aggregate.Min(&db.Animal.Id),
//	})
func Min(t any) *query.Min {
	return &query.Min{Field: t}
}

// Aggregate Sum uses database aggregate to sum the values on the target
//
// Is used with [query.Sum] as argument for select
//
// # Example
//
//	goe.Select[struct {
//		Sum query.Sum
//	}](&struct {
//		Sum *query.Sum
//	}{
//		Sum: aggregate.Sum(&db.Animal.Id),
//	})
func Sum(t any) *query.Sum {
	return &query.Sum{Field: t}
}
