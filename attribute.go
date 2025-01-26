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

func (p *pk) BuildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(p.selectName)
}

func (a *att) BuildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(a.selectName)
}

func (m *manyToOne) BuildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(m.selectName)
}

func (o *oneToOne) BuildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(o.selectName)
}

func (p *pk) BuildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.inserts = append(b.inserts, p)
	}
	b.returning = b.driver.Returning([]byte(p.attributeName))
	b.structPkName = p.structAttributeName
}

func (p *pk) WriteAttributeInsert(b *builder) {
	b.sql.WriteString(p.attributeName)
	b.attrNames = append(b.attrNames, p.structAttributeName)
}

func (a *att) BuildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, a)
}

func (a *att) WriteAttributeInsert(b *builder) {
	b.sql.WriteString(a.attributeName)
	b.attrNames = append(b.attrNames, a.structAttributeName)
}

func (m *manyToOne) BuildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, m)
}

func (m *manyToOne) WriteAttributeInsert(b *builder) {
	b.sql.WriteString(m.attributeName)
	b.attrNames = append(b.attrNames, m.structAttributeName)
}

func (o *oneToOne) BuildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, o)
}

func (o *oneToOne) WriteAttributeInsert(b *builder) {
	b.sql.WriteString(o.attributeName)
	b.attrNames = append(b.attrNames, o.structAttributeName)
}

func (p *pk) BuildAttributeUpdate(b *builder) {
	if !p.autoIncrement {
		b.attrNames = append(b.attrNames, p.attributeName)
		b.structColumns = append(b.structColumns, p.structAttributeName)
	}
}

func (a *att) BuildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, a.attributeName)
	b.structColumns = append(b.structColumns, a.structAttributeName)
}

func (m *manyToOne) BuildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, m.attributeName)
	b.structColumns = append(b.structColumns, m.structAttributeName)
}

func (o *oneToOne) BuildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, o.attributeName)
	b.structColumns = append(b.structColumns, o.structAttributeName)
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
