package goe

import (
	"maps"
	"reflect"
	"time"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type builder struct {
	query          model.Query
	modelStart     time.Time
	pkFieldId      int //insert
	fields         []field
	fieldsSelect   []fieldSelect
	fieldIds       []int           //insert and update
	joins          []enum.JoinType //select
	joinsArgs      []field         //select
	sets           []set
	whereArguments int
	tables         map[int]bool
	filter         *model.Where
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
	b.tables = make(map[int]bool)
}

func (b *builder) buildSelectJoins(join enum.JoinType, fields []field) {
	j := len(b.joinsArgs)
	b.joinsArgs = append(b.joinsArgs, make([]field, 2)...)
	b.joins = append(b.joins, join)
	b.joinsArgs[j] = fields[0]
	b.joinsArgs[j+1] = fields[1]
}

func (b *builder) buildSqlSelect() {
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
	b.query.Tables = make([]model.Table, 1)
	b.query.Tables[0] = model.Table{Schema: b.fields[0].schema(), Name: b.fields[0].table()}
	b.buildWhere()
	b.query.Header.ModelBuild = time.Since(b.modelStart)
}

func (b *builder) buildWhere() {
	if b.filter != nil && b.query.Where != nil {
		b.query.Where = &model.Where{
			Operator:        enum.And,
			Type:            enum.LogicalWhere,
			FirstOperation:  b.query.Where,
			SecondOperation: b.filter,
		}
	} else if b.filter != nil {
		b.query.Where = b.filter
	}

	if b.query.Where == nil {
		return
	}
	b.query.WhereIndex = len(b.query.Arguments) - b.whereArguments + 1
}

func (b *builder) buildTables() {
	if len(b.joins) != 0 {
		b.query.Joins = make([]model.Join, len(b.joins))
		tables := maps.Clone(b.tables)
		if !tables[b.joinsArgs[0].getTableId()] {
			b.query.Tables = append(b.query.Tables, model.Table{Schema: b.joinsArgs[0].schema(), Name: b.joinsArgs[0].table()})
			tables[b.joinsArgs[0].getTableId()] = true
		}
		c := 1
		for i := range b.joins {
			buildJoins(i, b.query.Joins, b.joins[i], b.joinsArgs[i+c-1], b.joinsArgs[i+c-1+1], tables)
			c++
		}
		return
	}
	if len(b.query.Tables) == 0 {
		b.query.Tables = append(b.query.Tables, model.Table{Schema: b.fieldsSelect[0].schema(), Name: b.fieldsSelect[0].table()})
	}
}

func buildJoins(i int, joins []model.Join, join enum.JoinType, f1, f2 field, tables map[int]bool) {
	if !tables[f1.getTableId()] {
		joins[i] = model.Join{
			Table:          model.Table{Schema: f1.schema(), Name: f1.table()},
			FirstArgument:  model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
			JoinOperation:  join,
			SecondArgument: model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}}

		tables[f1.getTableId()] = true
		return
	}
	joins[i] = model.Join{
		Table:          model.Table{Schema: f2.schema(), Name: f2.table()},
		FirstArgument:  model.JoinArgument{Table: f2.table(), Name: f2.getAttributeName()},
		JoinOperation:  join,
		SecondArgument: model.JoinArgument{Table: f1.table(), Name: f1.getAttributeName()}}

	tables[f2.getTableId()] = true
}

func (b *builder) buildInsert() {
	b.fieldIds = make([]int, 0, len(b.fields))
	b.query.Attributes = make([]model.Attribute, 0, len(b.fields))

	b.query.Tables = make([]model.Table, 1)
	b.query.Tables[0] = model.Table{Schema: b.fields[0].schema(), Name: b.fields[0].table()}
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
	b.buildWhere()
	b.query.Header.ModelBuild = time.Since(b.modelStart)
}

func (b *builder) buildSets() {
	b.query.Attributes = make([]model.Attribute, len(b.sets))
	b.query.Tables = make([]model.Table, 1)
	b.query.Tables[0] = model.Table{Schema: b.sets[0].attribute.schema(), Name: b.sets[0].attribute.table()}
	b.query.Arguments = make([]any, len(b.sets))

	for i := range b.sets {
		b.query.Attributes[i] = model.Attribute{Name: b.sets[i].attribute.getAttributeName()}
		b.query.Arguments[i] = b.sets[i].value
	}
}
