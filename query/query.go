package query

import (
	"fmt"

	"github.com/go-goe/goe/enum"
)

type Function[T any] struct {
	Field *T
	Type  enum.FunctionType
	Value T
}

func (f *Function[T]) Scan(src any) error {
	v, ok := src.(T)
	if !ok {
		return fmt.Errorf("error scan function")
	}

	f.Value = v
	return nil
}

func (f Function[T]) GetValue() any {
	return f.Value
}

func (f Function[T]) GetType() enum.FunctionType {
	return f.Type
}

type Count struct {
	Field any
	Value int64
}

func (c *Count) Scan(src any) error {
	v, ok := src.(int64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	c.Value = v
	return nil
}

func (c Count) Aggregate() enum.AggregateType {
	return enum.CountAggregate
}

type Avg struct {
	Field any
	Value float64
}

func (a *Avg) Scan(src any) error {
	v, ok := src.(float64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	a.Value = v
	return nil
}

func (a Avg) Aggregate() enum.AggregateType {
	return enum.AvgAggregate
}

type Max struct {
	Field any
	Value float64
}

func (a *Max) Scan(src any) error {
	v, ok := src.(float64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	a.Value = v
	return nil
}

func (a Max) Aggregate() enum.AggregateType {
	return enum.MaxAggregate
}

type Min struct {
	Field any
	Value float64
}

func (a *Min) Scan(src any) error {
	v, ok := src.(float64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	a.Value = v
	return nil
}

func (a Min) Aggregate() enum.AggregateType {
	return enum.MinAggregate
}

type Sum struct {
	Field any
	Value float64
}

func (a *Sum) Scan(src any) error {
	v, ok := src.(float64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	a.Value = v
	return nil
}

func (a Sum) Aggregate() enum.AggregateType {
	return enum.SumAggregate
}
