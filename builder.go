package goe

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/olauro/goe/query"
)

var ErrInvalidWhere = errors.New("goe: invalid where operation. try sending a pointer as parameter")
var ErrNoMatchesTables = errors.New("don't have any relationship")
var ErrNotManyToMany = errors.New("don't have a many to many relationship")

type builder struct {
	sql          *strings.Builder
	query        Query
	driver       Driver
	pkFieldId    int    //insert
	returning    []byte //insert
	inserts      []field
	froms        []byte
	fields       []field
	fieldsSelect []fieldSelect
	argsAny      []any
	fieldIds     []int //insert and update
	orderBy      string
	limit        uint     //select
	offset       uint     //select
	joins        []string //select
	joinsArgs    []field  //select
	tables       []int
	brs          []query.Operator
}

func createBuilder(d Driver, t uint) *builder {
	return &builder{
		sql:    &strings.Builder{},
		query:  Query{Type: t},
		driver: d,
	}
}

func (b *builder) buildSelect() {
	b.query.Attributes = make([]Attribute, 0, len(b.fieldsSelect))

	b.sql.Write(b.driver.Select())
	b.sql.WriteByte(' ')

	len := len(b.fieldsSelect)
	if len == 0 {
		return
	}

	for i := range b.fieldsSelect[:len-1] {
		b.fieldsSelect[i].buildAttributeSelect(b)
		b.sql.WriteByte(',')
	}

	b.fieldsSelect[len-1].buildAttributeSelect(b)
}

func (b *builder) buildSelectJoins(join string, fields []field) {
	j := len(b.joinsArgs)
	b.joinsArgs = append(b.joinsArgs, make([]field, 2)...)
	b.tables = append(b.tables, make([]int, 1)...)
	b.joins = append(b.joins, join)
	b.joinsArgs[j] = fields[0]
	b.joinsArgs[j+1] = fields[1]
}

func (b *builder) buildPage() {
	if b.limit != 0 {
		b.query.Limit = b.limit
		b.sql.WriteString(fmt.Sprintf(" LIMIT %v", b.limit))
	}
	if b.offset != 0 {
		b.query.Offset = b.offset
		b.sql.WriteString(fmt.Sprintf(" OFFSET %v", b.offset))
	}
}

func (b *builder) buildSqlSelect() (err error) {
	b.buildSelect()
	err = b.buildTables()
	if err != nil {
		return err
	}
	err = b.buildWhere()
	b.sql.WriteString(b.orderBy)
	b.buildPage()
	return err
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

func (b *builder) buildSqlUpdate(v reflect.Value) (err error) {
	b.buildUpdate()
	b.buildSet(v)
	err = b.buildWhere()
	return err
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
	b.query.WhereOperations = make([]Where, 0, len(b.brs))

	b.sql.WriteByte('\n')
	b.sql.Write(b.driver.Where())
	b.sql.WriteByte(' ')
	argsCount := len(b.argsAny) + 1
	for _, op := range b.brs {
		switch v := op.(type) {
		case query.Operation:
			v.ValueFlag = fmt.Sprintf("$%v", argsCount)
			b.sql.WriteString(v.Operation())
			b.argsAny = append(b.argsAny, v.Value)

			b.query.WhereOperations = append(b.query.WhereOperations, Where{
				Attribute: Attribute{
					Name:         v.Attribute,
					Table:        v.Table,
					FunctionType: v.Function,
				},
				Value:     v.Value,
				Operator:  v.Operator,
				ValueFlag: v.ValueFlag,
				Type:      OperationWhere,
			})
			argsCount++
		case query.OperationArg:
			b.query.WhereOperations = append(b.query.WhereOperations, Where{
				Attribute: Attribute{
					Name:  v.Op.Attribute,
					Table: v.Op.Table,
				},
				Value:     v.Op.Value,
				Operator:  v.Op.Operator,
				ValueFlag: v.Op.ValueFlag,
				Type:      OperationArgumentWhere,
			})

			b.sql.WriteString(v.Operation())
		case query.OperationIs:
			b.query.WhereOperations = append(b.query.WhereOperations, Where{
				Attribute: Attribute{
					Name:  v.Attribute,
					Table: v.Table,
				},
				Operator: v.Operator,
				Type:     OperationIsWhere,
			})

			b.sql.WriteString(v.Operation())
		case query.Logical:
			b.query.WhereOperations = append(b.query.WhereOperations, Where{
				Operator: v.Operator,
				Type:     LogicalWhere,
			})

			b.sql.WriteString(v.Operation())
		}
	}
	return nil
}

func (b *builder) buildTables() (err error) {
	if len(b.joins) != 0 {
		b.query.Joins = make([]Join, 0, len(b.joins))
	}
	b.sql.Write(b.driver.From())
	b.sql.Write(b.froms)
	c := 1
	for i := range b.joins {
		err = buildJoins(b, b.joins[i], b.sql, b.joinsArgs[i+c-1], b.joinsArgs[i+c-1+1], b.tables, i+1)
		if err != nil {
			return err
		}
		c++
	}
	return nil
}

func buildJoins(b *builder, join string, sql *strings.Builder, f1, f2 field, tables []int, tableIndice int) error {
	sql.WriteByte('\n')
	if !slices.Contains(tables, f2.getTableId()) {
		sql.WriteString(join)
		sql.WriteString("on(")
		sql.WriteString(f1.getSelect())
		sql.WriteByte('=')
		sql.WriteString(f2.getSelect())
		sql.WriteByte(')')

		b.query.Joins = append(b.query.Joins, Join{
			Table:          f2.table(),
			FirstArgument:  JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
			JoinOperation:  join,
			SecondArgument: JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}})

		tables[tableIndice] = f2.getTableId()
		return nil
	}
	sql.WriteString(join)
	sql.WriteString("on(")
	sql.WriteString(f1.getSelect())
	sql.WriteByte('=')
	sql.WriteString(f2.getSelect())
	sql.WriteByte(')')

	b.query.Joins = append(b.query.Joins, Join{
		Table:          f1.table(),
		FirstArgument:  JoinArgument{Table: f1.table(), Name: f1.getAttributeName()},
		JoinOperation:  join,
		SecondArgument: JoinArgument{Table: f2.table(), Name: f2.getAttributeName()}})

	tables[tableIndice] = f1.getTableId()
	return nil
}

func (b *builder) buildInsert() {
	b.sql.Write(b.driver.Insert())

	b.fieldIds = make([]int, 0, len(b.fields))
	b.query.Attributes = make([]Attribute, 0, len(b.fields))

	f := b.fields[0]
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = f.table()
	b.sql.WriteByte('(')
	for i := range b.fields {
		b.fields[i].buildAttributeInsert(b)
	}

	b.inserts[0].writeAttributeInsert(b)
	for _, f := range b.inserts[1:] {
		b.sql.WriteByte(',')
		f.writeAttributeInsert(b)
	}

	b.sql.WriteByte(')')
	b.sql.Write(b.driver.Values())
}

func (b *builder) buildValues(value reflect.Value) int {
	b.sql.WriteByte('(')
	b.argsAny = make([]any, 0, len(b.fieldIds))

	c := 2
	b.sql.WriteString("$1")
	b.argsAny = append(b.argsAny, value.Field(b.fieldIds[0]).Interface())

	a := b.fieldIds[1:]
	for i := range a {
		b.sql.WriteByte(',')
		b.sql.WriteString(fmt.Sprintf("$%v", c))
		b.argsAny = append(b.argsAny, value.Field(a[i]).Interface())
		c++
	}
	b.sql.WriteByte(')')
	if b.returning != nil {
		b.sql.Write(b.returning)
	}
	//TODO use Arguments
	b.query.Arguments = b.argsAny
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func (b *builder) buildBatchValues(value reflect.Value) int {
	b.argsAny = make([]any, 0, len(b.fieldIds))

	c := 1
	buildBatchValues(value.Index(0), b, &c)
	c++
	for j := 1; j < value.Len(); j++ {
		b.sql.WriteByte(',')
		buildBatchValues(value.Index(j), b, &c)
		c++
	}
	if b.returning != nil {
		b.sql.Write(b.returning)
	}
	b.query.BatchSizeQuery = value.Len()
	b.query.SizeArguments = len(b.fieldIds)
	return b.pkFieldId

}

func buildBatchValues(value reflect.Value, b *builder, c *int) {
	b.sql.WriteByte('(')
	b.sql.WriteString(fmt.Sprintf("$%v", *c))
	b.argsAny = append(b.argsAny, value.Field(b.fieldIds[0]).Interface())

	a := b.fieldIds[1:]
	for i := range a {
		b.sql.WriteByte(',')
		b.sql.WriteString(fmt.Sprintf("$%v", *c+1))
		b.argsAny = append(b.argsAny, value.Field(a[i]).Interface())
		*c++
	}
	b.sql.WriteByte(')')
	b.query.Arguments = b.argsAny
}

func (b *builder) buildUpdate() {
	b.sql.Write(b.driver.Update())

	b.fieldIds = make([]int, 0, len(b.fields))
	b.query.Attributes = make([]Attribute, 0, len(b.fields))
	b.query.Tables = make([]string, 1)
	b.query.Tables[0] = b.fields[0].table()

	b.sql.Write(b.driver.Set())
	b.fields[0].buildAttributeUpdate(b)

	a := b.fields[1:]
	for i := range a {
		a[i].buildAttributeUpdate(b)
	}
}

func (b *builder) buildSet(value reflect.Value) {
	b.argsAny = make([]any, 0, len(b.fieldIds))
	var c uint16 = 1
	buildSetField(value.Field(b.fieldIds[0]), b.fields[0].getAttributeName(), b, c)

	for i := 1; i < len(b.fieldIds); i++ {
		b.sql.WriteByte(',')
		c++
		buildSetField(value.Field(b.fieldIds[i]), b.fields[i].getAttributeName(), b, c)
	}
	b.query.Arguments = b.argsAny
}

func buildSetField(valueField reflect.Value, attributeName string, b *builder, c uint16) {
	b.query.Attributes = append(b.query.Attributes, Attribute{Name: attributeName})
	b.sql.WriteString(fmt.Sprintf("= $%v", c))
	b.argsAny = append(b.argsAny, valueField.Interface())
	c++
}
