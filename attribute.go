package goe

import (
	"fmt"
	"reflect"

	"github.com/olauro/goe/utils"
)

type oneToOne struct {
	primaryKey bool
	attributeStrings
}

func (o *oneToOne) getDb() *DB {
	return o.db
}

func (o *oneToOne) isPrimaryKey() bool {
	return o.primaryKey
}

func (o *oneToOne) table() []byte {
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
	primaryKey bool
	attributeStrings
}

func (m *manyToOne) getDb() *DB {
	return m.db
}

func (m *manyToOne) isPrimaryKey() bool {
	return m.primaryKey
}

func (m *manyToOne) table() []byte {
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

func (p *pk) getDb() *DB {
	return p.db
}

func (p *pk) isPrimaryKey() bool {
	return true
}

func (p *pk) table() []byte {
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

func (a *att) getDb() *DB {
	return a.db
}

func (a *att) isPrimaryKey() bool {
	return false
}

func (a *att) table() []byte {
	return a.tableBytes
}

func createAtt(db *DB, attributeName string, tableBytes []byte, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(db, tableBytes, attributeName, d)}
}

func (p *pk) buildAttributeSelect(b *builder) {
	b.sql.WriteString(p.selectName)
}

func (a *att) buildAttributeSelect(b *builder) {
	b.sql.WriteString(a.selectName)
}

func (m *manyToOne) buildAttributeSelect(b *builder) {
	b.sql.WriteString(m.selectName)
}

func (o *oneToOne) buildAttributeSelect(b *builder) {
	b.sql.WriteString(o.selectName)
}

func (p *pk) buildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.inserts = append(b.inserts, p)
	}
	b.returning = b.driver.Returning([]byte(p.attributeName))
	b.structPkName = p.structAttributeName
}

func (p *pk) writeAttributeInsert(b *builder) {
	b.sql.WriteString(p.attributeName)
	b.attrNames = append(b.attrNames, p.structAttributeName)
}

func (a *att) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, a)
}

func (a *att) writeAttributeInsert(b *builder) {
	b.sql.WriteString(a.attributeName)
	b.attrNames = append(b.attrNames, a.structAttributeName)
}

func (m *manyToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, m)
}

func (m *manyToOne) writeAttributeInsert(b *builder) {
	b.sql.WriteString(m.attributeName)
	b.attrNames = append(b.attrNames, m.structAttributeName)
}

func (o *oneToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, o)
}

func (o *oneToOne) writeAttributeInsert(b *builder) {
	b.sql.WriteString(o.attributeName)
	b.attrNames = append(b.attrNames, o.structAttributeName)
}

func (p *pk) buildAttributeUpdate(b *builder) {
	if !p.autoIncrement {
		b.attrNames = append(b.attrNames, p.attributeName)
		b.structColumns = append(b.structColumns, p.structAttributeName)
	}
}

func (a *att) buildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, a.attributeName)
	b.structColumns = append(b.structColumns, a.structAttributeName)
}

func (m *manyToOne) buildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, m.attributeName)
	b.structColumns = append(b.structColumns, m.structAttributeName)
}

func (o *oneToOne) buildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, o.attributeName)
	b.structColumns = append(b.structColumns, o.structAttributeName)
}

func (p *pk) getSelect() string {
	return p.selectName
}

func (a *att) getSelect() string {
	return a.selectName
}

func (m *manyToOne) getSelect() string {
	return m.selectName
}

func (o *oneToOne) getSelect() string {
	return o.selectName
}

type aggregate struct {
	selectName string
	db         *DB
}

func (a *aggregate) buildAttributeSelect(b *builder) {
	//TODO: update to write bytes
	b.sql.WriteString(a.selectName)
}

func (a *aggregate) getDb() *DB {
	return a.db
}
