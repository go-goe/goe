package goe

import (
	"fmt"
	"reflect"
	"strings"
)

type builder struct {
	sql            *strings.Builder
	args           []string
	argsAny        []any
	structColumns  []string          //select and update
	attrNames      []string          //insert and update
	targetFksNames map[string]string //insert and update
	joins          []string
	brs            []operator
	table          string
	tablesPk       []*pk
}

func createBuilder() *builder {
	return &builder{
		sql: &strings.Builder{}}
}

type statement struct {
	keyword     string
	allowCopies bool
	tip         int8
}

func createStatement(k string, t int8) *statement {
	return &statement{keyword: k, tip: t}
}

func (b *builder) buildSelect(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("SELECT ")

	b.structColumns = make([]string, len(b.args))
	b.tablesPk = make([]*pk, 1)

	for i := range b.args[:len(b.args)-1] {
		addrMap[b.args[i]].buildAttributeSelect(b, i)
		b.sql.WriteByte(44)
	}

	addrMap[b.args[len(b.args)-1]].buildAttributeSelect(b, len(b.args)-1)
	b.sql.WriteString(" FROM ")
	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
}

func (b *builder) buildSelectJoins(addrMap map[string]field, join string, argsJoins []string) {
	b.tablesPk = append(b.tablesPk, make([]*pk, 2)...)
	c := len(b.tablesPk) - 2
	b.joins = append(b.joins, join)
	b.tablesPk[c] = addrMap[argsJoins[0]].getPrimaryKey()
	b.tablesPk[c+1] = addrMap[argsJoins[1]].getPrimaryKey()
}

func (b *builder) buildSqlSelect() {
	b.buildTables()
	b.buildWhere()
	b.sql.WriteByte(59)
}

func (b *builder) buildSqlUpdate() {
	b.buildWhere()
	b.sql.WriteByte(59)
}

func (b *builder) buildSqlDelete() {
	b.buildWhere()
	b.sql.WriteByte(59)
}

func (b *builder) buildSqlUpdateIn() {
	b.buildWhereIn()
	b.sql.WriteByte(59)
}

func (b *builder) buildWhere() {
	if len(b.brs) == 0 {
		return
	}
	b.sql.WriteByte(10)
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1
	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			v.setValueFlag(fmt.Sprintf("$%v", argsCount))
			b.sql.WriteString(v.operation())
			b.argsAny = append(b.argsAny, v.value)
			argsCount++
		case simpleOperator:
			b.sql.WriteString(v.operation())
		}
	}
}

func (b *builder) buildWhereIn() {
	if len(b.brs) == 0 {
		return
	}
	b.sql.WriteByte(10)
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1

	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			st := buildWhereIn(b.tablesPk, v.pk, argsCount, v)
			if st != "" {
				b.sql.WriteString(st)
				b.argsAny = append(b.argsAny, v.value)
				argsCount++
			}
		case simpleOperator:
			b.sql.WriteString(v.operation())
		}
	}
}

func buildWhereIn(pks []*pk, brPk *pk, argsCount int, v complexOperator) string {
	for i := range pks {
		mtm := brPk.fks[pks[i].table]
		if mtm != nil {
			if mtmValue, ok := mtm.(*manyToMany); ok {
				v.setValueFlag(fmt.Sprintf("$%v", argsCount))
				v.setArgument(mtmValue.ids[brPk.table].attributeName)
				return v.operation()
			}
		}
	}
	return ""
}

func (b *builder) buildTables() {
	if len(b.tablesPk) > 1 {
		c := 1
		for i := range b.joins {
			buildJoins(b.tablesPk[i+c], b.tablesPk[i+c+1], b.joins[i], b.sql, i+c+1, b.tablesPk)
			c++
		}
	}
}

func buildJoins(pk1 *pk, pk2 *pk, join string, sql *strings.Builder, i int, pks []*pk) {
	switch fk := pk1.fks[pk2.table].(type) {
	case *manyToOne:
		if fk.hasMany {
			sql.WriteByte(10)
			sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk2.table, pk1.selectName, fk.selectName))
		} else {
			sql.WriteByte(10)
			sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk1.table, fk.selectName, pk2.selectName))
		}
	case *manyToMany:
		for c := range pks[:i] {
			//switch pks if pk2 is priority
			if pks[c].table == pk2.table {
				pk2 = pk1
				pk1 = pks[c]
				break
			}
		}
		sql.WriteByte(10)
		sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, fk.table, pk1.selectName, fk.ids[pk1.table].selectName))
		sql.WriteByte(10)
		sql.WriteString(fmt.Sprintf(
			"%v %v on (%v = %v)",
			join,
			pk2.table, fk.ids[pk2.table].selectName,
			pk2.selectName))
	}
}

func (b *builder) buildInsert(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.targetFksNames = make(map[string]string)
	b.attrNames = make([]string, 0, len(b.args))
	b.tablesPk = make([]*pk, 1)

	f := addrMap[b.args[0]]
	b.sql.WriteString(f.getPrimaryKey().table)
	b.sql.WriteString(" (")
	b.tablesPk[0] = f.getPrimaryKey()
	f.buildAttributeInsert(b)
	if !f.getPrimaryKey().autoIncrement {
		b.sql.WriteByte(44)
	}

	l := len(b.args[1:]) - 1

	a := b.args[1:]
	for i := range a {
		addrMap[a[i]].buildAttributeInsert(b)
		if i != l {
			b.sql.WriteByte(44)
		}
	}
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
}

func (b *builder) buildValues(value reflect.Value) string {
	b.sql.WriteString("(")
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 2
	b.sql.WriteString("$1")
	buildValueField(value.FieldByName(b.attrNames[0]), b.attrNames[0], b)
	a := b.attrNames[1:]
	for i := range a {
		b.sql.WriteByte(44)
		b.sql.WriteString(fmt.Sprintf("$%v", c))
		buildValueField(value.FieldByName(a[i]), a[i], b)
		c++
	}
	pk := b.tablesPk[0]
	b.sql.WriteString(") ")
	b.sql.WriteString("RETURNING ")
	st := createStatement(pk.attributeName, 0)
	st.allowCopies = true
	b.sql.WriteString(pk.attributeName)
	b.sql.WriteByte(59)
	return pk.structAttributeName

}

func buildValueField(valueField reflect.Value, fieldName string, b *builder) {
	switch valueField.Kind() {
	case reflect.Struct:
		if valueField.Type().Name() != "Time" {
			b.argsAny = append(b.argsAny, valueField.FieldByName(b.targetFksNames[fieldName]).Interface())
			return
		}
	case reflect.Pointer:
		if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
			b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(b.targetFksNames[fieldName]).Interface())
			return
		}
	}
	b.argsAny = append(b.argsAny, valueField.Interface())
}

func (b *builder) buildInsertIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildValuesIn() {
	pk1 := b.tablesPk[0]
	pk2 := b.tablesPk[1]

	mtm := pk2.fks[b.table]
	if mtm == nil {
		return
	}

	mtmValue := mtm.(*manyToMany)
	b.sql.WriteString(mtmValue.table)
	b.sql.WriteString(" (")
	b.sql.WriteString(mtmValue.ids[pk1.table].attributeName)
	b.sql.WriteString(",")
	b.sql.WriteString(mtmValue.ids[pk2.table].attributeName)
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
	b.sql.WriteString("($1,$2);")
}

func (b *builder) buildUpdate(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.targetFksNames = make(map[string]string)
	b.structColumns = make([]string, 0, len(b.args))
	b.attrNames = make([]string, 0, len(b.args))

	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
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
		b.sql.WriteByte(44)
		c++
		buildSetField(value.FieldByName(s[i]), a[i], b, c)
	}
}

func buildSetField(valueField reflect.Value, fieldName string, b *builder, c uint16) {
	switch valueField.Kind() {
	case reflect.Struct:
		if valueField.Type().Name() == "Time" {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.Interface())
			c++
			return
		}
		if !valueField.FieldByName(b.targetFksNames[fieldName]).IsZero() {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.FieldByName(b.targetFksNames[fieldName]).Interface())
			c++
		}
		return
	case reflect.Pointer:
		if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(b.targetFksNames[fieldName]).Interface())
			c++
			return
		}
	}
	b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
	b.argsAny = append(b.argsAny, valueField.Interface())
	c++
}

func (b *builder) buildUpdateIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.attrNames = make([]string, 0, len(b.args))
	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildSetIn() {
	pk2 := b.tablesPk[1]

	mtm := pk2.fks[b.table]
	if mtm == nil {
		return
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		b.sql.WriteString(" SET ")
		b.sql.WriteString(fmt.Sprintf("%v = $1", mtmValue.ids[pk2.table].attributeName))
	}
}

func (b *builder) buildDelete(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")
	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
}

func (b *builder) buildDeleteIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")

	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildSqlDeleteIn() {
	mtm := b.tablesPk[1].fks[b.table]
	if mtm == nil {
		//TODO: add error
		return
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		b.buildWhereIn()
		b.sql.WriteByte(59)
	}
	//TODO: add error
}
