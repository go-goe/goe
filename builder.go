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

type Builder struct {
	Sql           *strings.Builder
	Driver        Driver
	StructPkName  string //insert
	Returning     []byte //insert
	Inserts       []Field
	Froms         []byte
	Args          []uintptr
	Aggregates    []aggregate
	ArgsAny       []any
	StructColumns []string //select and update
	AttrNames     []string //insert and update
	OrderBy       string
	Limit         uint     //select
	Offset        uint     //select
	Joins         []string //select
	JoinsArgs     []Field  //select
	Tables        []string //select TODO: update all table names to a int ID
	Brs           []wh.Operator
}

func CreateBuilder(d Driver) *Builder {
	return &Builder{
		Sql:    &strings.Builder{},
		Driver: d,
	}
}

func (b *Builder) BuildSelect(addrMap map[uintptr]Field) {
	b.Sql.Write(b.Driver.Select())

	if len(b.Aggregates) > 0 {
		b.buildAggregates()
	}

	lenArgs := len(b.Args)
	if lenArgs == 0 {
		return
	}

	b.StructColumns = make([]string, lenArgs)

	for i := range b.Args[:lenArgs-1] {
		addrMap[b.Args[i]].BuildAttributeSelect(b, i)
		b.Sql.WriteByte(',')
	}

	addrMap[b.Args[lenArgs-1]].BuildAttributeSelect(b, lenArgs-1)
}

func (b *Builder) buildAggregates() {
	for i := range b.Aggregates[:len(b.Aggregates)-1] {
		b.Sql.WriteString(b.Aggregates[i].String())
		b.Sql.WriteByte(',')
	}
	b.Sql.WriteString(b.Aggregates[len(b.Aggregates)-1].String())
}

func (b *Builder) BuildSelectJoins(addrMap map[uintptr]Field, join string, ArgsJoins []uintptr) {
	j := len(b.JoinsArgs)
	b.JoinsArgs = append(b.JoinsArgs, make([]Field, 2)...)
	b.Tables = append(b.Tables, make([]string, 1)...)
	b.Joins = append(b.Joins, join)
	b.JoinsArgs[j] = addrMap[ArgsJoins[0]]
	b.JoinsArgs[j+1] = addrMap[ArgsJoins[1]]
}

func (b *Builder) buildPage() {
	if b.Limit != 0 {
		b.Sql.WriteString(fmt.Sprintf(" LIMIT %v", b.Limit))
	}
	if b.Offset != 0 {
		b.Sql.WriteString(fmt.Sprintf(" OFFSET %v", b.Offset))
	}
}

func (b *Builder) BuildSqlSelect() (err error) {
	err = b.buildTables()
	if err != nil {
		return err
	}
	err = b.buildWhere()
	b.Sql.WriteString(b.OrderBy)
	b.buildPage()
	b.Sql.WriteByte(';')
	return err
}

func (b *Builder) BuildSqlUpdate() (err error) {
	err = b.buildWhere()
	b.Sql.WriteByte(';')
	return err
}

func (b *Builder) BuildSqlDelete() (err error) {
	err = b.buildWhere()
	b.Sql.WriteByte(';')
	return err
}

func (b *Builder) buildWhere() error {
	if len(b.Brs) == 0 {
		return nil
	}
	b.Sql.WriteByte('\n')
	b.Sql.WriteString("WHERE ")
	ArgsCount := len(b.ArgsAny) + 1
	for _, op := range b.Brs {
		switch v := op.(type) {
		case wh.Operation:
			v.ValueFlag = fmt.Sprintf("$%v", ArgsCount)
			b.Sql.WriteString(v.Operation())
			b.ArgsAny = append(b.ArgsAny, v.Value)
			ArgsCount++
		default:
			b.Sql.WriteString(v.Operation())
		}
	}
	return nil
}

func (b *Builder) buildTables() (err error) {
	b.Sql.Write(b.Driver.From())
	b.Sql.Write(b.Froms)
	c := 1
	for i := range b.Joins {
		err = buildJoins(b.Joins[i], b.Sql, b.JoinsArgs[i+c-1], b.JoinsArgs[i+c-1+1], b.Tables, i+1)
		if err != nil {
			return err
		}
		c++
	}
	return nil
}

func buildJoins(join string, Sql *strings.Builder, f1, f2 Field, Tables []string, tableIndice int) error {
	Sql.WriteByte('\n')
	if !slices.Contains(Tables, string(f2.Table())) {
		Sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, string(f2.Table()), f1.GetSelect(), f2.GetSelect()))
		Tables[tableIndice] = string(f2.Table())
		return nil
	}
	//TODO: update this to write
	Sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, string(f1.Table()), f1.GetSelect(), f2.GetSelect()))
	Tables[tableIndice] = string(f1.Table())
	return nil
}

func (b *Builder) BuildInsert(addrMap map[uintptr]Field) {
	//TODO: Set a drive type to share stm
	b.Sql.WriteString("INSERT ")
	b.Sql.WriteString("INTO ")

	b.AttrNames = make([]string, 0, len(b.Args))

	f := addrMap[b.Args[0]]
	b.Sql.Write(f.Table())
	b.Sql.WriteString(" (")
	for i := range b.Args {
		addrMap[b.Args[i]].BuildAttributeInsert(b)
	}

	b.Inserts[0].WriteAttributeInsert(b)
	for _, f := range b.Inserts[1:] {
		b.Sql.WriteByte(',')
		f.WriteAttributeInsert(b)
	}

	b.Sql.WriteString(") ")
	b.Sql.WriteString("VALUES ")
}

func (b *Builder) BuildValues(value reflect.Value) string {
	b.Sql.WriteByte(40)
	b.ArgsAny = make([]any, 0, len(b.AttrNames))

	c := 2
	b.Sql.WriteString("$1")
	buildValueField(value.FieldByName(b.AttrNames[0]), b)
	a := b.AttrNames[1:]
	for i := range a {
		b.Sql.WriteByte(',')
		b.Sql.WriteString(fmt.Sprintf("$%v", c))
		buildValueField(value.FieldByName(a[i]), b)
		c++
	}
	b.Sql.WriteByte(')')
	if b.Returning != nil {
		b.Sql.Write(b.Returning)
	}
	return b.StructPkName

}

func (b *Builder) BuildBatchValues(value reflect.Value) string {
	b.ArgsAny = make([]any, 0, len(b.AttrNames))

	c := 1
	buildBatchValues(value.Index(0), b, &c)
	c++
	for j := 1; j < value.Len(); j++ {
		b.Sql.WriteByte(',')
		buildBatchValues(value.Index(j), b, &c)
		c++
	}
	if b.Returning != nil {
		b.Sql.Write(b.Returning)
	}
	return b.StructPkName

}

func buildBatchValues(value reflect.Value, b *Builder, c *int) {
	b.Sql.WriteByte(40)
	b.Sql.WriteString(fmt.Sprintf("$%v", *c))
	buildValueField(value.FieldByName(b.AttrNames[0]), b)
	a := b.AttrNames[1:]
	for i := range a {
		b.Sql.WriteByte(',')
		b.Sql.WriteString(fmt.Sprintf("$%v", *c+1))
		buildValueField(value.FieldByName(a[i]), b)
		*c++
	}
	b.Sql.WriteByte(')')
}

func buildValueField(valueField reflect.Value, b *Builder) {
	b.ArgsAny = append(b.ArgsAny, valueField.Interface())
}

func (b *Builder) BuildUpdate(addrMap map[uintptr]Field) {
	//TODO: Set a drive type to share stm
	b.Sql.WriteString("UPDATE ")

	b.StructColumns = make([]string, 0, len(b.Args))
	b.AttrNames = make([]string, 0, len(b.Args))

	b.Sql.Write(addrMap[b.Args[0]].Table())
	b.Sql.WriteString(" SET ")
	addrMap[b.Args[0]].BuildAttributeUpdate(b)

	a := b.Args[1:]
	for i := range a {
		addrMap[a[i]].BuildAttributeUpdate(b)
	}
}

func (b *Builder) BuildSet(value reflect.Value) {
	b.ArgsAny = make([]any, 0, len(b.AttrNames))
	var c uint16 = 1
	buildSetField(value.FieldByName(b.StructColumns[0]), b.AttrNames[0], b, c)

	a := b.AttrNames[1:]
	s := b.StructColumns[1:]
	for i := range a {
		b.Sql.WriteByte(',')
		c++
		buildSetField(value.FieldByName(s[i]), a[i], b, c)
	}
}

func buildSetField(valueField reflect.Value, FieldName string, b *Builder, c uint16) {
	b.Sql.WriteString(fmt.Sprintf("%v = $%v", FieldName, c))
	b.ArgsAny = append(b.ArgsAny, valueField.Interface())
	c++
}

func (b *Builder) BuildDelete(addrMap map[uintptr]Field) {
	//TODO: Set a drive type to share stm
	b.Sql.WriteString("DELETE FROM ")
	b.Sql.Write(addrMap[b.Args[0]].Table())
}
