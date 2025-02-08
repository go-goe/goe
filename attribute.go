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

func (o *oneToOne) getTableId() uint {
	return o.tableId
}

func (o *oneToOne) table() []byte {
	return o.tableBytes
}

func (o *oneToOne) getAttributeName() []byte {
	return o.attributeName
}

func createOneToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, tableId uint, Driver Driver, prefix, fieldName string) *oneToOne {
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
		fieldName,
		tableId,
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

func (m *manyToOne) getTableId() uint {
	return m.tableId
}

func (m *manyToOne) table() []byte {
	return m.tableBytes
}

func (m *manyToOne) getAttributeName() []byte {
	return m.attributeName
}

func createManyToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, tableId uint, Driver Driver, prefix, fieldName string) *manyToOne {
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
		fieldName,
		tableId,
		Driver,
	)
	return mto
}

type attributeStrings struct {
	db                  *DB
	tableId             uint
	tableBytes          []byte
	selectName          string
	attributeName       []byte
	structAttributeName string
}

func createAttributeStrings(db *DB, table []byte, attributeName string, tableId uint, Driver Driver) attributeStrings {
	return attributeStrings{
		db:                  db,
		tableBytes:          table,
		tableId:             tableId,
		selectName:          fmt.Sprintf("%v.%v", string(table), Driver.KeywordHandler(utils.ColumnNamePattern(attributeName))),
		attributeName:       []byte(Driver.KeywordHandler(utils.ColumnNamePattern(attributeName))),
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

func (p *pk) getTableId() uint {
	return p.tableId
}

func (p *pk) table() []byte {
	return p.tableBytes
}

func (p *pk) getAttributeName() []byte {
	return p.attributeName
}

func createPk(db *DB, table []byte, attributeName string, autoIncrement bool, tableId uint, Driver Driver) *pk {
	//TODO:: Check this utils
	table = []byte(Driver.KeywordHandler(utils.TableNamePattern(string(table))))
	return &pk{
		attributeStrings: createAttributeStrings(db, table, attributeName, tableId, Driver),
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

func (a *att) getTableId() uint {
	return a.tableId
}

func (a *att) table() []byte {
	return a.tableBytes
}

func (a *att) getAttributeName() []byte {
	return a.attributeName
}

func createAtt(db *DB, attributeName string, tableBytes []byte, tableId uint, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(db, tableBytes, attributeName, tableId, d)}
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
	b.returning = b.driver.Returning(p.attributeName)
	b.structPkName = p.structAttributeName
}

func (p *pk) writeAttributeInsert(b *builder) {
	b.sql.Write(p.attributeName)
	b.structColumns = append(b.structColumns, p.structAttributeName)
}

func (a *att) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, a)
}

func (a *att) writeAttributeInsert(b *builder) {
	b.sql.Write(a.attributeName)
	b.structColumns = append(b.structColumns, a.structAttributeName)
}

func (m *manyToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, m)
}

func (m *manyToOne) writeAttributeInsert(b *builder) {
	b.sql.Write(m.attributeName)
	b.structColumns = append(b.structColumns, m.structAttributeName)
}

func (o *oneToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, o)
}

func (o *oneToOne) writeAttributeInsert(b *builder) {
	b.sql.Write(o.attributeName)
	b.structColumns = append(b.structColumns, o.structAttributeName)
}

func (p *pk) buildAttributeUpdate(b *builder) {
	if !p.autoIncrement {
		b.structColumns = append(b.structColumns, p.structAttributeName)
	}
}

func (a *att) buildAttributeUpdate(b *builder) {
	b.structColumns = append(b.structColumns, a.structAttributeName)
}

func (m *manyToOne) buildAttributeUpdate(b *builder) {
	b.structColumns = append(b.structColumns, m.structAttributeName)
}

func (o *oneToOne) buildAttributeUpdate(b *builder) {
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
