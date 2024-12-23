package goe

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/olauro/goe/wh"
)

var ErrInvalidWhere = errors.New("goe: invalid where operation. try sending a pointer as parameter")
var ErrNoMatchesTables = errors.New("don't have any relationship")
var ErrNotManyToMany = errors.New("don't have a many to many relationship")

type builder struct {
	sql           *strings.Builder
	driver        Driver
	structPkName  string //insert
	returning     []byte //insert
	froms         []byte
	args          []uintptr
	aggregates    []aggregate
	argsAny       []any
	structColumns []string //select and update
	attrNames     []string //insert and update
	orderBy       string
	limit         uint
	offset        uint
	joins         []string //select
	joinsArgs     []field  //select
	tables        []string //select TODO: update all table names to a int ID
	brs           []operator
}

func createBuilder(d Driver) *builder {
	return &builder{
		sql:    &strings.Builder{},
		driver: d,
	}
}

func (b *builder) buildSelect(addrMap map[uintptr]field) {
	b.sql.Write(b.driver.Select())

	if len(b.aggregates) > 0 {
		b.buildAggregates()
	}

	lenArgs := len(b.args)
	if lenArgs == 0 {
		return
	}

	b.structColumns = make([]string, lenArgs)

	for i := range b.args[:lenArgs-1] {
		addrMap[b.args[i]].buildAttributeSelect(b, i)
		b.sql.WriteByte(',')
	}

	addrMap[b.args[lenArgs-1]].buildAttributeSelect(b, lenArgs-1)
}

func (b *builder) buildAggregates() {
	for i := range b.aggregates[:len(b.aggregates)-1] {
		b.sql.WriteString(b.aggregates[i].String())
		b.sql.WriteByte(',')
	}
	b.sql.WriteString(b.aggregates[len(b.aggregates)-1].String())
}

func (b *builder) buildSelectJoins(addrMap map[uintptr]field, join string, argsJoins []uintptr) {
	j := len(b.joinsArgs)
	b.joinsArgs = append(b.joinsArgs, make([]field, 2)...)
	b.tables = append(b.tables, make([]string, 1)...)
	b.joins = append(b.joins, join)
	b.joinsArgs[j] = addrMap[argsJoins[0]]
	b.joinsArgs[j+1] = addrMap[argsJoins[1]]
}

func (b *builder) buildPage() {
	if b.limit != 0 {
		b.sql.WriteString(fmt.Sprintf(" LIMIT %v", b.limit))
	}
	if b.offset != 0 {
		b.sql.WriteString(fmt.Sprintf(" OFFSET %v", b.offset))
	}
}

func (b *builder) buildSqlSelect() (err error) {
	err = b.buildTables()
	if err != nil {
		return err
	}
	err = b.buildWhere()
	b.sql.WriteString(b.orderBy)
	b.buildPage()
	b.sql.WriteByte(';')
	return err
}

func (b *builder) buildSqlUpdate() (err error) {
	err = b.buildWhere()
	b.sql.WriteByte(';')
	return err
}

func (b *builder) buildSqlDelete() (err error) {
	err = b.buildWhere()
	b.sql.WriteByte(';')
	return err
}

func (b *builder) buildWhere() error {
	if len(b.brs) == 0 {
		return nil
	}
	b.sql.WriteByte('\n')
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1
	for _, op := range b.brs {
		switch v := op.(type) {
		case wh.Operation:
			v.ValueFlag = fmt.Sprintf("$%v", argsCount)
			b.sql.WriteString(v.Operation())
			b.argsAny = append(b.argsAny, v.Value)
			argsCount++
		default:
			b.sql.WriteString(v.Operation())
		}
	}
	return nil
}

func (b *builder) buildTables() (err error) {
	b.sql.Write(b.driver.From())
	b.sql.Write(b.froms)
	c := 1
	for i := range b.joins {
		err = buildJoins(b.joins[i], b.sql, b.joinsArgs[i+c-1], b.joinsArgs[i+c-1+1], b.tables, i+1)
		if err != nil {
			return err
		}
		c++
	}
	return nil
}

func buildJoins(join string, sql *strings.Builder, f1, f2 field, tables []string, tableIndice int) error {
	sql.WriteByte('\n')
	if !slices.Contains(tables, string(f2.table())) {
		sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, string(f2.table()), f1.getSelect(), f2.getSelect()))
		tables[tableIndice] = string(f2.table())
		return nil
	}
	//TODO: update this to write
	sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, string(f1.table()), f1.getSelect(), f2.getSelect()))
	tables[tableIndice] = string(f1.table())
	return nil
}

func (b *builder) buildInsert(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.attrNames = make([]string, 0, len(b.args))

	f := addrMap[b.args[0]]
	b.sql.Write(f.table())
	b.sql.WriteString(" (")
	f.buildAttributeInsert(b)
	if !f.getPrimaryKey().autoIncrement {
		b.sql.WriteByte(',')
	}

	l := len(b.args[1:]) - 1

	a := b.args[1:]
	for i := range a {
		addrMap[a[i]].buildAttributeInsert(b)
		if i != l {
			b.sql.WriteByte(',')
		}
	}
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
}

func (b *builder) buildValues(value reflect.Value) string {
	b.sql.WriteByte(40)
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 2
	b.sql.WriteString("$1")
	buildValueField(value.FieldByName(b.attrNames[0]), b)
	a := b.attrNames[1:]
	for i := range a {
		b.sql.WriteByte(',')
		b.sql.WriteString(fmt.Sprintf("$%v", c))
		buildValueField(value.FieldByName(a[i]), b)
		c++
	}
	b.sql.WriteByte(')')
	if b.returning != nil {
		b.sql.Write(b.returning)
	}
	return b.structPkName

}

func (b *builder) buildBatchValues(value reflect.Value) string {
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 1
	buildBatchValues(value.Index(0), b, &c)
	c++
	for j := 1; j < value.Len(); j++ {
		b.sql.WriteByte(',')
		buildBatchValues(value.Index(j), b, &c)
		c++
	}
	//pk := b.tablesPk[0]
	if b.returning != nil {
		b.sql.Write(b.returning)
	}
	return b.structPkName

}

func buildBatchValues(value reflect.Value, b *builder, c *int) {
	b.sql.WriteByte(40)
	b.sql.WriteString(fmt.Sprintf("$%v", *c))
	buildValueField(value.FieldByName(b.attrNames[0]), b)
	a := b.attrNames[1:]
	for i := range a {
		b.sql.WriteByte(',')
		b.sql.WriteString(fmt.Sprintf("$%v", *c+1))
		buildValueField(value.FieldByName(a[i]), b)
		*c++
	}
	b.sql.WriteByte(')')
}

func buildValueField(valueField reflect.Value, b *builder) {
	b.argsAny = append(b.argsAny, valueField.Interface())
}

func (b *builder) buildUpdate(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.structColumns = make([]string, 0, len(b.args))
	b.attrNames = make([]string, 0, len(b.args))

	b.sql.Write(addrMap[b.args[0]].table())
	b.sql.WriteString(" SET ")
	addrMap[b.args[0]].buildAttributeUpdate(b)

	a := b.args[1:]
	for i := range a {
		addrMap[a[i]].buildAttributeUpdate(b)
	}
}

func (b *builder) buildSet(value reflect.Value) {
	b.argsAny = make([]any, 0, len(b.attrNames))
	var c uint16 = 1
	buildSetField(value.FieldByName(b.structColumns[0]), b.attrNames[0], b, c)

	a := b.attrNames[1:]
	s := b.structColumns[1:]
	for i := range a {
		b.sql.WriteByte(',')
		c++
		buildSetField(value.FieldByName(s[i]), a[i], b, c)
	}
}

func buildSetField(valueField reflect.Value, fieldName string, b *builder, c uint16) {
	b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
	b.argsAny = append(b.argsAny, valueField.Interface())
	c++
}

func (b *builder) buildDelete(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")
	b.sql.Write(addrMap[b.args[0]].table())
}
