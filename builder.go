package goe

import (
	"reflect"
	"time"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type builder struct {
	query        model.Query
	modelStart   time.Time
	pkFieldId    int //insert
	fields       []field
	fieldsSelect []fieldSelect
	fieldIds     []int           //insert and update
	joins        []enum.JoinType //select
	joinsArgs    []field         //select
	tables       map[int]int
	brs          []model.Operation
	sets         []set
}

type set struct {
	attribute field
	value     any
}

func createBuilder(typeQuery enum.QueryType) builder {
	return builder{
		query:      model.Query{Type: typeQuery},
		modelStart: time.Now(),
	}
}

func (b *builder) buildSelect() {
	b.query.Attributes = make([]model.Attribute, len(b.fieldsSelect))
	for i := range b.fieldsSelect {
		b.fieldsSelect[i].buildAttributeSelect(b.query.Attributes, i)
	}
}

func (b *builder) buildSelectJoins(join enum.JoinType, fields []field) {
	j := len(b.joinsArgs)
	b.joinsArgs = append(b.joinsArgs, make([]field, 2)...)
	b.joins = append(b.joins, join)
	b.joinsArgs[j] = fields[0]
	b.joinsArgs[j+1] = fields[1]
}

func (b *builder) buildSqlSelect() {
	b.buildSelect()
	b.buildTables()
	b.buildWhere()
	b.query.Header.ModelBuild = time.Since(b.modelStart)
}

func (b *builder) buildSqlInsert(v reflect.Value) (pkFieldId int) {
	b.buildInsert()
	pkFieldId = b.buildValues(v)
	b.query.Header.ModelBuild = time.Since(b.modelStart)
	return pkFieldId
}

func (b *builder) buildSqlInsertBatch(v reflect.Value) (pkFieldId int) {
	b.buildInsert()
	pkFieldId = b.buildBatchValues(v)
	b.query.Header.ModelBuild = time.Since(b.modelStart)
	return pkFieldId
}

func (b *builder) buildSqlDelete() {
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.fields[0].table()
	b.buildWhere()
	b.query.Header.ModelBuild = time.Since(b.modelStart)
}

func (b *builder) buildWhere() {
	if len(b.brs) == 0 {
		return
	}
	b.query.WhereOperations = make([]model.Where, len(b.brs))

	b.query.WhereIndex = len(b.query.Arguments) + 1
	for i, v := range b.brs {
		switch v.Type {
		case enum.OperationWhere:
			b.query.Arguments = append(b.query.Arguments, v.Value.GetValue())

			b.query.WhereOperations[i] = model.Where{
				Attribute: model.Attribute{
					Name:         v.Attribute,
					Table:        v.Table,
					FunctionType: v.Function,
				},
				Operator: v.Operator,
				Type:     v.Type,
			}
		case enum.OperationAttributeWhere:
			b.query.WhereOperations[i] = model.Where{
				Attribute: model.Attribute{
					Name:  v.Attribute,
					Table: v.Table,
				},
				Operator:       v.Operator,
				AttributeValue: model.Attribute{Name: v.AttributeValue, Table: v.AttributeValueTable},
				Type:           v.Type,
			}
		case enum.OperationIsWhere:
			b.query.WhereOperations[i] = model.Where{
				Attribute: model.Attribute{
					Name:  v.Attribute,
					Table: v.Table,
				},
				Operator: v.Operator,
				Type:     v.Type,
			}
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
					b.query.Arguments = append(b.query.Arguments, valueOf.Index(i).Interface())
					where.SizeIn++
				}
			case reflect.Array:
				for i := range valueOf.Len() {
					b.query.Arguments = append(b.query.Arguments, valueOf.Index(i).Interface())
					where.SizeIn++
				}
			default:
				if modelQuery, ok := valueOf.Interface().(model.Query); ok {
					where.QueryIn = &modelQuery
				}
			}

			b.query.WhereOperations[i] = where
		case enum.LogicalWhere:
			b.query.WhereOperations[i] = model.Where{
				Operator: v.Operator,
				Type:     v.Type,
			}

		}
	}
}

func (b *builder) buildTables() {
	if len(b.joins) != 0 {
		b.tables[b.joinsArgs[0].getTableId()] = 1
		b.query.Tables = append(b.query.Tables, b.joinsArgs[0].table())

		b.query.Joins = make([]model.Join, len(b.joins))
		c := 1
		for i := range b.joins {
			buildJoins(i, b.query.Joins, b.joins[i], b.joinsArgs[i+c-1], b.joinsArgs[i+c-1+1], b.tables)
			c++
		}
		return
	}
	b.tables[b.fieldsSelect[0].getTableId()] = 1
	b.query.Tables = append(b.query.Tables, b.fieldsSelect[0].table())

	for i := range b.brs {
		if b.brs[i].TableId != 0 && b.tables[b.brs[i].TableId] == 0 {
			b.tables[b.brs[i].TableId] = 1
			b.query.Tables = append(b.query.Tables, b.brs[i].Table)
		}
		if b.brs[i].AttributeTableId != 0 && b.tables[b.brs[i].AttributeTableId] == 0 {
			b.tables[b.brs[i].AttributeTableId] = 1
			b.query.Tables = append(b.query.Tables, b.brs[i].AttributeValueTable)
		}
	}
}

func buildJoins(i int, joins []model.Join, join enum.JoinType, f1, f2 field, tables map[int]int) {
	if tables[f1.getTableId()] == 1 {
		joins[i] = model.Join{
			Table:          f2.table(),
			FirstArgument:  model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
			JoinOperation:  join,
			SecondArgument: model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}}

		tables[f2.getTableId()] = 1
		return
	}
	joins[i] = model.Join{
		Table:          f1.table(),
		FirstArgument:  model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
		JoinOperation:  join,
		SecondArgument: model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}}

	tables[f1.getTableId()] = 1
}

func (b *builder) buildInsert() {
	b.fieldIds = make([]int, 0, len(b.fields))
	b.query.Attributes = make([]model.Attribute, 0, len(b.fields))

	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.fields[0].table()
	for i := range b.fields {
		b.fields[i].buildAttributeInsert(b)
	}
}

func (b *builder) buildValues(value reflect.Value) int {
	b.query.Arguments = make([]any, len(b.fieldIds))

	for c, i := range b.fieldIds {
		b.query.Arguments[c] = value.Field(i).Interface()
	}
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func (b *builder) buildBatchValues(value reflect.Value) int {
	b.query.Arguments = make([]any, len(b.fieldIds)*value.Len())

	c := 0
	for j := 0; j < value.Len(); j++ {
		c = buildBatchValues(value.Index(j), b, c)
	}
	b.query.BatchSizeQuery = value.Len()
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func buildBatchValues(value reflect.Value, b *builder, c int) int {
	for _, i := range b.fieldIds {
		b.query.Arguments[c] = value.Field(i).Interface()
		c++
	}
	return c
}

func (b *builder) buildUpdate() {
	b.buildSets()
	b.buildWhere()
	b.query.Header.ModelBuild = time.Since(b.modelStart)
}

func (b *builder) buildSets() {
	b.query.Attributes = make([]model.Attribute, len(b.sets))
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.sets[0].attribute.table()
	b.query.Arguments = make([]any, len(b.sets))

	for i := range b.sets {
		b.query.Attributes[i] = model.Attribute{Name: b.sets[i].attribute.getAttributeName()}
		b.query.Arguments[i] = b.sets[i].value
	}
}
