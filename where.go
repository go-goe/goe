package goe

import (
	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/query/where"
)

type customWhere struct {
	modelWhere model.Where
}

func (cw customWhere) getModel() model.Where {
	return cw.modelWhere
}

func (cw customWhere) And(c customWhere) customWhere {
	return customWhere{modelWhere: where.And(cw.modelWhere, c.modelWhere)}
}

func (cw customWhere) Or(c customWhere) customWhere {
	return customWhere{modelWhere: where.Or(cw.modelWhere, c.modelWhere)}
}

func equalsWhere[T any](v T, f field) customWhere {
	return customWhere{modelWhere: model.Where{Arg: f, Value: valueOperation{value: v}, Operator: enum.Equals, Type: enum.OperationWhere}}
}

func equalsNilWhere[T any](v *T, f field) customWhere {
	var m model.Where
	if v == nil {
		m = model.Where{Arg: f, Operator: enum.Is, Type: enum.OperationIsWhere}
	} else {
		m = model.Where{Arg: f, Value: valueOperation{value: v}, Operator: enum.Equals, Type: enum.OperationWhere}
	}
	return customWhere{modelWhere: m}
}

// TODO: Check this
type valueOperation struct {
	value any
}

func (vo valueOperation) GetValue() any {
	if result, ok := vo.value.(model.ValueOperation); ok {
		return result.GetValue()
	}
	return vo.value
}
