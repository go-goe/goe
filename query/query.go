package query

import (
	"fmt"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
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

func (f Function[T]) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:        b.Table,
		Name:         b.Name,
		FunctionType: f.Type,
	}
}

func (f Function[T]) GetField() any {
	return f.Field
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

func (c Count) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.CountAggregate,
	}
}

func (c Count) GetField() any {
	return c.Field
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

func (a Avg) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.AvgAggregate,
	}
}

func (a Avg) GetField() any {
	return a.Field
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

func (m Max) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.MaxAggregate,
	}
}

func (m Max) GetField() any {
	return m.Field
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

func (m Min) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.MinAggregate,
	}
}

func (m Min) GetField() any {
	return m.Field
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

func (s Sum) Attribute(b model.Body) model.Attribute {
	return model.Attribute{
		Table:         b.Table,
		Name:          b.Name,
		AggregateType: enum.SumAggregate,
	}
}

func (s Sum) GetField() any {
	return s.Field
}
