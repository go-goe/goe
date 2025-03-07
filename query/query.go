package query

import (
	"fmt"

	"github.com/olauro/goe/enum"
)

type Joins interface {
	FirstArg() any
	Join() string
	SecondArg() any
}

type Operation struct {
	Type                enum.WhereType
	Arg                 any
	Value               ValueOperation
	Operator            string
	Attribute           string
	Table               string
	Function            enum.FunctionType
	AttributeValue      string
	AttributeValueTable string
}

type ValueOperation interface {
	GetValue() any
}

type Function[T any] struct {
	Field *T
	Type  enum.FunctionType
	Value T
}

func (f *Function[string]) Scan(src any) error {
	v, ok := src.(string)
	if !ok {
		return fmt.Errorf("error scan function")
	}

	f.Value = v
	return nil
}

func (f Function[T]) GetValue() any {
	return f.Value
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

func (c *Count) Aggregate(s string) string {
	return fmt.Sprintf("COUNT(%v)", s)
}
