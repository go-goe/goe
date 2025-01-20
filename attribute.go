package goe

import (
	"fmt"
	"reflect"

	"github.com/olauro/goe/utils"
)

type oneToOne struct {
	isPrimaryKey bool
	attributeStrings
}

func (o *oneToOne) GetDb() *DB {
	return o.db
}

func (o *oneToOne) IsPrimaryKey() bool {
	return o.isPrimaryKey
}

func (o *oneToOne) Table() []byte {
	return o.tableBytes
}

func createOneToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, Driver Driver, prefix string) *oneToOne {
	mto := new(oneToOne)
	targetPks := primaryKeys(typeOf)
	count := 0
	for i := range targetPks {
		if targetPks[i].Name == prefix {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto.attributeStrings = createAttributeStrings(
		db,
		[]byte(Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))),
		prefix+typeOf.Name(),
		Driver,
	)
	return mto
}

type manyToOne struct {
	isPrimaryKey bool
	attributeStrings
}

func (m *manyToOne) GetDb() *DB {
	return m.db
}

func (m *manyToOne) IsPrimaryKey() bool {
	return m.isPrimaryKey
}

func (m *manyToOne) Table() []byte {
	return m.tableBytes
}

func createManyToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, Driver Driver, prefix string) *manyToOne {
	mto := new(manyToOne)
	targetPks := primaryKeys(typeOf)
	count := 0
	for i := range targetPks {
		if targetPks[i].Name == prefix {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto.attributeStrings = createAttributeStrings(
		db,
		[]byte(Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))),
		prefix+typeOf.Name(),
		Driver,
	)
	return mto
}

type attributeStrings struct {
	db                  *DB
	tableBytes          []byte
	selectName          string
	attributeName       string
	structAttributeName string
}

func createAttributeStrings(db *DB, table []byte, attributeName string, Driver Driver) attributeStrings {
	return attributeStrings{
		db:                  db,
		tableBytes:          table,
		selectName:          fmt.Sprintf("%v.%v", string(table), Driver.KeywordHandler(utils.ColumnNamePattern(attributeName))),
		attributeName:       Driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		structAttributeName: attributeName,
	}
}

type pk struct {
	autoIncrement bool
	attributeStrings
}

func (p *pk) GetDb() *DB {
	return p.db
}

func (p *pk) IsPrimaryKey() bool {
	return true
}

func (p *pk) Table() []byte {
	return p.tableBytes
}

func createPk(db *DB, table []byte, attributeName string, autoIncrement bool, Driver Driver) *pk {
	//TODO:: Check this utils
	table = []byte(Driver.KeywordHandler(utils.TableNamePattern(string(table))))
	return &pk{
		attributeStrings: createAttributeStrings(db, table, attributeName, Driver),
		autoIncrement:    autoIncrement}
}

type att struct {
	attributeStrings
}

func (a *att) GetDb() *DB {
	return a.db
}

func (a *att) IsPrimaryKey() bool {
	return false
}

func (a *att) Table() []byte {
	return a.tableBytes
}

func createAtt(db *DB, attributeName string, tableBytes []byte, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(db, tableBytes, attributeName, d)}
}

func (p *pk) BuildAttributeSelect(b *Builder, i int) {
	b.Sql.WriteString(p.selectName)
	b.StructColumns[i] = p.structAttributeName
}

func (a *att) BuildAttributeSelect(b *Builder, i int) {
	b.Sql.WriteString(a.selectName)
	b.StructColumns[i] = a.structAttributeName
}

func (m *manyToOne) BuildAttributeSelect(b *Builder, i int) {
	b.Sql.WriteString(m.selectName)
	b.StructColumns[i] = m.structAttributeName
}

func (o *oneToOne) BuildAttributeSelect(b *Builder, i int) {
	b.Sql.WriteString(o.selectName)
	b.StructColumns[i] = o.structAttributeName
}

func (p *pk) BuildAttributeInsert(b *Builder) {
	if !p.autoIncrement {
		b.Inserts = append(b.Inserts, p)
	}
	b.Returning = b.Driver.Returning([]byte(p.attributeName))
	b.StructPkName = p.structAttributeName
}

func (p *pk) WriteAttributeInsert(b *Builder) {
	b.Sql.WriteString(p.attributeName)
	b.AttrNames = append(b.AttrNames, p.structAttributeName)
}

func (a *att) BuildAttributeInsert(b *Builder) {
	b.Inserts = append(b.Inserts, a)
}

func (a *att) WriteAttributeInsert(b *Builder) {
	b.Sql.WriteString(a.attributeName)
	b.AttrNames = append(b.AttrNames, a.structAttributeName)
}

func (m *manyToOne) BuildAttributeInsert(b *Builder) {
	b.Inserts = append(b.Inserts, m)
}

func (m *manyToOne) WriteAttributeInsert(b *Builder) {
	b.Sql.WriteString(m.attributeName)
	b.AttrNames = append(b.AttrNames, m.structAttributeName)
}

func (o *oneToOne) BuildAttributeInsert(b *Builder) {
	b.Inserts = append(b.Inserts, o)
}

func (o *oneToOne) WriteAttributeInsert(b *Builder) {
	b.Sql.WriteString(o.attributeName)
	b.AttrNames = append(b.AttrNames, o.structAttributeName)
}

func (p *pk) BuildAttributeUpdate(b *Builder) {
	if !p.autoIncrement {
		b.AttrNames = append(b.AttrNames, p.attributeName)
		b.StructColumns = append(b.StructColumns, p.structAttributeName)
	}
}

func (a *att) BuildAttributeUpdate(b *Builder) {
	b.AttrNames = append(b.AttrNames, a.attributeName)
	b.StructColumns = append(b.StructColumns, a.structAttributeName)
}

func (m *manyToOne) BuildAttributeUpdate(b *Builder) {
	b.AttrNames = append(b.AttrNames, m.attributeName)
	b.StructColumns = append(b.StructColumns, m.structAttributeName)
}

func (o *oneToOne) BuildAttributeUpdate(b *Builder) {
	b.AttrNames = append(b.AttrNames, o.attributeName)
	b.StructColumns = append(b.StructColumns, o.structAttributeName)
}

func (p *pk) GetSelect() string {
	return p.selectName
}

func (a *att) GetSelect() string {
	return a.selectName
}

func (m *manyToOne) GetSelect() string {
	return m.selectName
}

func (o *oneToOne) GetSelect() string {
	return o.selectName
}

type aggregate struct {
	function string
	field    Field
}

func createAggregate(function string, f Field) aggregate {
	return aggregate{function: function, field: f}
}

func (a aggregate) String() string {
	return fmt.Sprintf("%v(%v)", a.function, a.field.GetSelect())
}
