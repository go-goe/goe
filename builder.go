package goe

import (
	"errors"
	"reflect"
	"slices"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
)

type builder struct {
	query        model.Query
	pkFieldId    int //insert
	inserts      []field
	fields       []field
	fieldsSelect []fieldSelect
	fieldIds     []int           //insert and update
	joins        []enum.JoinType //select
	joinsArgs    []field         //select
	tables       []int
	brs          []model.Operation
	sets         []set
}

type set struct {
	attribute field
	value     any
}

func createBuilder(typeQuery enum.QueryType) builder {
	return builder{
		query: model.Query{Type: typeQuery},
	}
}

func (b *builder) buildSelect() {
	b.query.Attributes = make([]model.Attribute, 0, len(b.fieldsSelect))

	len := len(b.fieldsSelect)
	if len == 0 {
		return
	}

	for i := range b.fieldsSelect[:len-1] {
		b.fieldsSelect[i].buildAttributeSelect(b)
	}

	b.fieldsSelect[len-1].buildAttributeSelect(b)
}

func (b *builder) buildSelectJoins(join enum.JoinType, fields []field) {
	j := len(b.joinsArgs)
	b.joinsArgs = append(b.joinsArgs, make([]field, 2)...)
	b.tables = append(b.tables, make([]int, 1)...)
	b.joins = append(b.joins, join)
	b.joinsArgs[j] = fields[0]
	b.joinsArgs[j+1] = fields[1]
}

func (b *builder) buildSqlSelect() (err error) {
	if b.query.Tables == nil {
		return errors.New("goe: none table specified, use From() function to specify")
	}
	b.buildSelect()
	err = b.buildTables()
	if err != nil {
		return err
	}
	return b.buildWhere()
}

func (b *builder) buildSqlInsert(v reflect.Value) (pkFieldId int) {
	b.buildInsert()
	pkFieldId = b.buildValues(v)
	return pkFieldId
}

func (b *builder) buildSqlInsertBatch(v reflect.Value) (pkFieldId int) {
	b.buildInsert()
	pkFieldId = b.buildBatchValues(v)
	return pkFieldId
}

func (b *builder) buildSqlDelete() (err error) {
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.fields[0].table()
	err = b.buildWhere()
	return err
}

func (b *builder) buildWhere() error {
	if len(b.brs) == 0 {
		return nil
	}
	b.query.WhereOperations = make([]model.Where, 0, len(b.brs))

	b.query.WhereIndex = len(b.query.Arguments) + 1
	for _, v := range b.brs {
		switch v.Type {
		case enum.OperationWhere:
			b.query.Arguments = append(b.query.Arguments, v.Value.GetValue())

			b.query.WhereOperations = append(b.query.WhereOperations, model.Where{
				Attribute: model.Attribute{
					Name:         v.Attribute,
					Table:        v.Table,
					FunctionType: v.Function,
				},
				Operator: v.Operator,
				Type:     v.Type,
			})
		case enum.OperationAttributeWhere:
			b.query.WhereOperations = append(b.query.WhereOperations, model.Where{
				Attribute: model.Attribute{
					Name:  v.Attribute,
					Table: v.Table,
				},
				Operator:       v.Operator,
				AttributeValue: model.Attribute{Name: v.AttributeValue, Table: v.AttributeValueTable},
				Type:           v.Type,
			})
		case enum.OperationIsWhere:
			b.query.WhereOperations = append(b.query.WhereOperations, model.Where{
				Attribute: model.Attribute{
					Name:  v.Attribute,
					Table: v.Table,
				},
				Operator: v.Operator,
				Type:     v.Type,
			})
		case enum.OperationInWhere:
			where := model.Where{
				Attribute: model.Attribute{
					Name:         v.Attribute,
					Table:        v.Table,
					FunctionType: v.Function,
				},
				Operator: v.Operator,
				Type:     v.Type,
			}

			valueOf := reflect.ValueOf(v.Value.GetValue())
			switch valueOf.Kind() {
			case reflect.Slice:
				for i := range valueOf.Len() {
					b.query.Arguments = append(b.query.Arguments, valueOf.Index(i))
					where.SizeIn++
				}
			case reflect.Array:
				for i := range valueOf.Len() {
					b.query.Arguments = append(b.query.Arguments, valueOf.Index(i))
					where.SizeIn++
				}
			default:
				if modelQuery, ok := valueOf.Interface().(*model.Query); ok {
					where.QueryIn = modelQuery
				}
			}

			b.query.WhereOperations = append(b.query.WhereOperations, where)
		case enum.LogicalWhere:
			b.query.WhereOperations = append(b.query.WhereOperations, model.Where{
				Operator: v.Operator,
				Type:     v.Type,
			})

		}
	}
	return nil
}

func (b *builder) buildTables() (err error) {
	if len(b.joins) != 0 {
		b.query.Joins = make([]model.Join, 0, len(b.joins))
	}
	c := 1
	for i := range b.joins {
		buildJoins(b, b.joins[i], b.joinsArgs[i+c-1], b.joinsArgs[i+c-1+1], b.tables, i+1)
		c++
	}
	return nil
}

func buildJoins(b *builder, join enum.JoinType, f1, f2 field, tables []int, tableIndice int) {
	if slices.Contains(tables, f1.getTableId()) {
		b.query.Joins = append(b.query.Joins, model.Join{
			Table:          f2.table(),
			FirstArgument:  model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
			JoinOperation:  join,
			SecondArgument: model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}})

		tables[tableIndice] = f2.getTableId()
		return
	}
	b.query.Joins = append(b.query.Joins, model.Join{
		Table:          f1.table(),
		FirstArgument:  model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
		JoinOperation:  join,
		SecondArgument: model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}})

	tables[tableIndice] = f1.getTableId()
}

func (b *builder) buildInsert() {

	b.fieldIds = make([]int, 0, len(b.fields))
	b.query.Attributes = make([]model.Attribute, 0, len(b.fields))

	f := b.fields[0]
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = f.table()
	for i := range b.fields {
		b.fields[i].buildAttributeInsert(b)
	}

	b.inserts[0].writeAttributeInsert(b)
	for _, f := range b.inserts[1:] {
		f.writeAttributeInsert(b)
	}

}

func (b *builder) buildValues(value reflect.Value) int {
	//update to index
	b.query.Arguments = make([]any, 0, len(b.fieldIds))

	c := 2
	b.query.Arguments = append(b.query.Arguments, value.Field(b.fieldIds[0]).Interface())

	a := b.fieldIds[1:]
	for i := range a {
		b.query.Arguments = append(b.query.Arguments, value.Field(a[i]).Interface())
		c++
	}
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func (b *builder) buildBatchValues(value reflect.Value) int {
	b.query.Arguments = make([]any, 0, len(b.fieldIds))

	c := 1
	buildBatchValues(value.Index(0), b, &c)
	c++
	for j := 1; j < value.Len(); j++ {
		buildBatchValues(value.Index(j), b, &c)
		c++
	}
	b.query.BatchSizeQuery = value.Len()
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func buildBatchValues(value reflect.Value, b *builder, c *int) {
	b.query.Arguments = append(b.query.Arguments, value.Field(b.fieldIds[0]).Interface())

	a := b.fieldIds[1:]
	for i := range a {
		b.query.Arguments = append(b.query.Arguments, value.Field(a[i]).Interface())
		*c++
	}
}

func (b *builder) buildUpdate() (err error) {
	if len(b.sets) == 0 {
		return errors.New("goe: can't update on empty set")
	}
	b.buildSets()
	return b.buildWhere()
}

func (b *builder) buildSets() {
	b.query.Attributes = make([]model.Attribute, 0, len(b.sets))
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.sets[0].attribute.table()
	b.query.Arguments = make([]any, 0, len(b.sets))

	for i := range b.sets {
		b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: b.sets[i].attribute.getAttributeName()})
		b.query.Arguments = append(b.query.Arguments, b.sets[i].value)
	}
}
