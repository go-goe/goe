package goe

import (
	"reflect"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/utils"
)

type oneToOne struct {
	attributeStrings
}

func (o oneToOne) scheme() *string {
	return o.schemeName
}

func (o oneToOne) getDb() *DB {
	return o.db
}

func (o oneToOne) isPrimaryKey() bool {
	return false
}

func (o oneToOne) getTableId() int {
	return o.tableId
}

func (o oneToOne) table() string {
	return o.tableName
}

func (o oneToOne) getAttributeName() string {
	return o.attributeName
}

func createOneToOne(b body, typeOf reflect.Type) any {
	mto := oneToOne{}
	targetPks := primaryKeys(typeOf)
	count := 0
	for i := range targetPks {
		if targetPks[i].Name == b.prefixName {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto.attributeStrings = createAttributeStrings(
		b.mapp.db,
		b.scheme,
		b.driver.KeywordHandler(utils.TableNamePattern(b.typeOf.Name())),
		b.fieldName,
		b.mapp.tableId,
		b.fieldId,
		b.driver,
	)
	return mto
}

type manyToOne struct {
	attributeStrings
}

func (m manyToOne) scheme() *string {
	return m.schemeName
}

func (m manyToOne) getDb() *DB {
	return m.db
}

func (m manyToOne) isPrimaryKey() bool {
	return false
}

func (m manyToOne) getTableId() int {
	return m.tableId
}

func (m manyToOne) table() string {
	return m.tableName
}

func (m manyToOne) getAttributeName() string {
	return m.attributeName
}

func createManyToOne(b body, typeOf reflect.Type) any {
	mto := manyToOne{}
	targetPks := primaryKeys(typeOf)
	count := 0
	for i := range targetPks {
		if targetPks[i].Name == b.prefixName {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto.attributeStrings = createAttributeStrings(
		b.mapp.db,
		b.scheme,
		b.driver.KeywordHandler(utils.TableNamePattern(b.typeOf.Name())),
		b.fieldName,
		b.mapp.tableId,
		b.fieldId,
		b.driver,
	)
	return mto
}

type attributeStrings struct {
	db            *DB
	schemeName    *string
	tableId       int
	tableName     string
	attributeName string
	fieldId       int
}

func createAttributeStrings(db *DB, scheme *string, table string, attributeName string, tableId, fieldId int, Driver Driver) attributeStrings {
	return attributeStrings{
		db:            db,
		tableName:     table,
		tableId:       tableId,
		fieldId:       fieldId,
		schemeName:    scheme,
		attributeName: Driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
	}
}

type pk struct {
	autoIncrement bool
	attributeStrings
}

func (p pk) scheme() *string {
	return p.schemeName
}

func (p pk) getDb() *DB {
	return p.db
}

func (p pk) isPrimaryKey() bool {
	return true
}

func (p pk) getTableId() int {
	return p.tableId
}

func (p pk) table() string {
	return p.tableName
}

func (p pk) getAttributeName() string {
	return p.attributeName
}

func createPk(db *DB, scheme *string, table string, attributeName string, autoIncrement bool, tableId, fieldId int, Driver Driver) pk {
	table = Driver.KeywordHandler(utils.TableNamePattern(table))
	return pk{
		attributeStrings: createAttributeStrings(db, scheme, table, attributeName, tableId, fieldId, Driver),
		autoIncrement:    autoIncrement}
}

type att struct {
	attributeStrings
}

func (a att) scheme() *string {
	return a.schemeName
}

func (a att) getDb() *DB {
	return a.db
}

func (a att) isPrimaryKey() bool {
	return false
}

func (a att) getTableId() int {
	return a.tableId
}

func (a att) table() string {
	return a.tableName
}

func (a att) getAttributeName() string {
	return a.attributeName
}

func createAtt(db *DB, attributeName string, scheme *string, table string, tableId, fieldId int, d Driver) att {
	return att{
		attributeStrings: createAttributeStrings(db, scheme, table, attributeName, tableId, fieldId, d)}
}

func (p pk) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table: p.tableName,
		Name:  p.attributeName,
	}
}

func (a att) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table: a.tableName,
		Name:  a.attributeName,
	}
}

func (m manyToOne) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table: m.tableName,
		Name:  m.attributeName,
	}
}

func (o oneToOne) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table: o.tableName,
		Name:  o.attributeName,
	}
}

func (p pk) buildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.fieldIds = append(b.fieldIds, p.fieldId)
		b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: p.getAttributeName()})
		return
	}
	b.query.ReturningId = &model.Attribute{Name: p.getAttributeName()}
	b.pkFieldId = p.fieldId
}

func (a att) buildAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, a.fieldId)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: a.getAttributeName()})
}

func (m manyToOne) buildAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, m.fieldId)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: m.getAttributeName()})
}

func (o oneToOne) buildAttributeInsert(b *builder) {
	b.fieldIds = append(b.fieldIds, o.fieldId)
	b.query.Attributes = append(b.query.Attributes, model.Attribute{Name: o.getAttributeName()})
}

func (p pk) getFieldId() int {
	return p.fieldId
}

func (a att) getFieldId() int {
	return a.fieldId
}

func (m manyToOne) getFieldId() int {
	return m.fieldId
}

func (o oneToOne) getFieldId() int {
	return o.fieldId
}

type aggregateResult struct {
	attributeName string
	tableName     string
	schemeName    *string
	aggregateType enum.AggregateType
	tableId       int
	db            *DB
}

func (a aggregateResult) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table:         a.tableName,
		Name:          a.attributeName,
		AggregateType: a.aggregateType}
}

func (a aggregateResult) scheme() *string {
	return a.schemeName
}

func (a aggregateResult) table() string {
	return a.tableName
}

func (a aggregateResult) getTableId() int {
	return a.tableId
}

func (a aggregateResult) getDb() *DB {
	return a.db
}

type functionResult struct {
	attributeName string
	tableName     string
	schemeName    *string
	functionType  enum.FunctionType
	tableId       int
	db            *DB
}

func (f functionResult) buildAttributeSelect(atts []model.Attribute, i int) {
	atts[i] = model.Attribute{
		Table:        f.tableName,
		Name:         f.attributeName,
		FunctionType: f.functionType}
}

func (f functionResult) scheme() *string {
	return f.schemeName
}

func (f functionResult) table() string {
	return f.tableName
}

func (f functionResult) getTableId() int {
	return f.tableId
}

func (f functionResult) getDb() *DB {
	return f.db
}
