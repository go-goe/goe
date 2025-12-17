package aggregate

import (
	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

// Aggregate Count uses database aggregate to make a count on the target.
//
// # Example
//
//	goe.Select[struct {
//		Count count
//	}](aggregate.Count(&db.Animal.Id))...
func Count(t any) *count {
	return &count{Field: t}
}

// Aggregate Avg uses database aggregate to get a average on the target.
//
// # Example
//
//	goe.Select[struct {
//		Avg float64
//	}](aggregate.Avg(&db.Exam.Result))...
func Avg(t any) *avg {
	return &avg{Field: t}
}

// Aggregate Max uses database aggregate to get the maximum value on the target.
//
// # Example
//
//	goe.Select[struct {
//		Max float64
//	}](aggregate.Max(&db.Exam.Result))...
func Max(t any) *max {
	return &max{Field: t}
}

// Aggregate Min uses database aggregate to get the minimum value on the target.
//
// # Example
//
//	goe.Select[struct {
//		Min float64
//	}](aggregate.Min(&db.Exam.Result))...
func Min(t any) *min {
	return &min{Field: t}
}

// Aggregate Sum uses database aggregate to sum the values on the target.
//
// # Example
//
//	goe.Select[struct {
//		Sum float64
//	}](aggregate.Sum(&db.Exam.Result))...
func Sum(t any) *sum {
	return &sum{Field: t}
}

type count struct {
	Field any
	Value int64
}

func (c count) Aggregate() enum.AggregateType {
	return enum.CountAggregate
}

func (c count) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.CountAggregate,
	}
}

func (c count) GetField() any {
	return c.Field
}

type avg struct {
	Field any
	Value float64
}

func (a avg) Aggregate() enum.AggregateType {
	return enum.AvgAggregate
}

func (a avg) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.AvgAggregate,
	}
}

func (a avg) GetField() any {
	return a.Field
}

type max struct {
	Field any
	Value float64
}

func (a max) Aggregate() enum.AggregateType {
	return enum.MaxAggregate
}

func (m max) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.MaxAggregate,
	}
}

func (m max) GetField() any {
	return m.Field
}

type min struct {
	Field any
	Value float64
}

func (a min) Aggregate() enum.AggregateType {
	return enum.MinAggregate
}

func (m min) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.MinAggregate,
	}
}

func (m min) GetField() any {
	return m.Field
}

type sum struct {
	Field any
	Value float64
}

func (a sum) Aggregate() enum.AggregateType {
	return enum.SumAggregate
}

func (s sum) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.SumAggregate,
	}
}

func (s sum) GetField() any {
	return s.Field
}
