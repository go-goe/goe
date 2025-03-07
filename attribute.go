package goe

import (
	"reflect"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
	"github.com/olauro/goe/utils"
)

type oneToOne struct {
	attributeStrings
}

func (o *oneToOne) getDb() *DB {
	return o.db
}

func (o *oneToOne) isPrimaryKey() bool {
	return false
}

func (o *oneToOne) getTableId() int {
	return o.tableId
}

func (o *oneToOne) table() string {
	return o.tableName
}

func (o *oneToOne) getAttributeName() string {
	return o.attributeName
}

func createOneToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, tableId, fieldId int, Driver Driver, prefix, fieldName string) *oneToOne {
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
		Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		fieldName,
		tableId,
		fieldId,
		Driver,
	)
	return mto
}

type manyToOne struct {
	attributeStrings
}

func (m *manyToOne) getDb() *DB {
	return m.db
}

func (m *manyToOne) isPrimaryKey() bool {
	return false
}

func (m *manyToOne) getTableId() int {
	return m.tableId
}

func (m *manyToOne) table() string {
	return m.tableName
}

func (m *manyToOne) getAttributeName() string {
	return m.attributeName
}

func createManyToOne(db *DB, typeOf reflect.Type, targetTypeOf reflect.Type, tableId, fieldId int, Driver Driver, prefix, fieldName string) *manyToOne {
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
		Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		fieldName,
		tableId,
		fieldId,
		Driver,
	)
	return mto
}

type attributeStrings struct {
	db            *DB
	tableId       int
	tableName     string
	attributeName string
	fieldId       int
}

func createAttributeStrings(db *DB, table string, attributeName string, tableId, fieldId int, Driver Driver) attributeStrings {
	return attributeStrings{
		db:            db,
		tableName:     table,
		tableId:       tableId,
		fieldId:       fieldId,
		attributeName: Driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
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

func (p *pk) getTableId() int {
	return p.tableId
}

func (p *pk) table() string {
	return p.tableName
}

func (p *pk) getAttributeName() string {
	return p.attributeName
}

func createPk(db *DB, table string, attributeName string, autoIncrement bool, tableId, fieldId int, Driver Driver) *pk {
	table = Driver.KeywordHandler(utils.TableNamePattern(table))
	return &pk{
		attributeStrings: createAttributeStrings(db, table, attributeName, tableId, fieldId, Driver),
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

func (a *att) getTableId() int {
	return a.tableId
}

func (a *att) table() string {
	return a.tableName
}

func (a *att) getAttributeName() string {
	return a.attributeName
}

func createAtt(db *DB, attributeName string, table string, tableId, fieldId int, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(db, table, attributeName, tableId, fieldId, d)}
}

func (p *pk) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table: p.tableName,
		Name:  p.attributeName,
	})
}

func (a *att) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table: a.tableName,
		Name:  a.attributeName,
	})
}

func (m *manyToOne) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table: m.tableName,
		Name:  m.attributeName,
	})
}

func (o *oneToOne) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table: o.tableName,
		Name:  o.attributeName,
	})
}

func (p *pk) buildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.inserts = append(b.inserts, p)
		b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: p.getAttributeName()})
		return
	}
	b.query.ReturningId = &model.Attribute{Name: p.getAttributeName()}
	b.pkFieldId = p.fieldId
}

func (p *pk) writeAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, p.fieldId)
}

func (a *att) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, a)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: a.getAttributeName()})
}

func (a *att) writeAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, a.fieldId)
}

func (m *manyToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, m)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: m.getAttributeName()})
}

func (m *manyToOne) writeAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, m.fieldId)
}

func (o *oneToOne) buildAttributeInsert(b *builder) {
	b.inserts = append(b.inserts, o)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: o.getAttributeName()})
}

func (o *oneToOne) writeAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, o.fieldId)
}

func (p *pk) buildAttributeUpdate(b *builder) {
	if !p.autoIncrement {
		b.fieldIds = append(b.fieldIds, p.fieldId)
	}
}

func (a *att) buildAttributeUpdate(b *builder) {
	b.fieldIds = append(b.fieldIds, a.fieldId)
}

func (m *manyToOne) buildAttributeUpdate(b *builder) {
	b.fieldIds = append(b.fieldIds, m.fieldId)
}

func (o *oneToOne) buildAttributeUpdate(b *builder) {
	b.fieldIds = append(b.fieldIds, o.fieldId)
}

func (p *pk) getFieldId() int {
	return p.fieldId
}

func (a *att) getFieldId() int {
	return a.fieldId
}

func (m *manyToOne) getFieldId() int {
	return m.fieldId
}

func (o *oneToOne) getFieldId() int {
	return o.fieldId
}

type aggregateResult struct {
	attributeName string
	table         string
	aggregateType enum.AggregateType
	db            *DB
}

func (a *aggregateResult) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table:         a.table,
		Name:          a.attributeName,
		AggregateType: a.aggregateType})
}

func (a *aggregateResult) getDb() *DB {
	return a.db
}

type functionResult struct {
	attributeName string
	table         string
	functionType  enum.FunctionType
	db            *DB
}

func (f *functionResult) buildAttributeSelect(b *builder) {
	b.query.Attributes = append(b.query.Attributes, model.Attribute{
		Table:        f.table,
		Name:         f.attributeName,
		FunctionType: f.functionType})
}

func (f *functionResult) getDb() *DB {
	return f.db
}
