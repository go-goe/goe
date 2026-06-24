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

func (o oneToOne) schema() *string {
	return o.schemaName
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
		b.schema,
		b.mapp.pks[0].tableName,
		b.fieldName,
		b.mapp.tableId,
		b.fieldId,
		b.driver,
	)
	return mto
}

type manyToOne struct {
	isDefault bool
	attributeStrings
}

func (m manyToOne) schema() *string {
	return m.schemaName
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
	mto.isDefault = getTagValue(b.valueOf.Type().Field(b.fieldId).Tag.Get("goe"), "default:") != ""
	mto.attributeStrings = createAttributeStrings(
		b.mapp.db,
		b.schema,
		b.mapp.pks[0].tableName,
		b.fieldName,
		b.mapp.tableId,
		b.fieldId,
		b.driver,
	)
	return mto
}

type attributeStrings struct {
	db            *DB
	schemaName    *string
	tableId       int
	tableName     string
	attributeName string
	fieldId       int
}

func createAttributeStrings(db *DB, schema *string, table string, attributeName string, tableId, fieldId int, Driver model.Driver) attributeStrings {
	return attributeStrings{
		db:            db,
		tableName:     table,
		tableId:       tableId,
		fieldId:       fieldId,
		schemaName:    schema,
		attributeName: Driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
	}
}

type pk struct {
	autoIncrement bool
	attributeStrings
}

func (p pk) schema() *string {
	return p.schemaName
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

func createPk(db *DB, schema *string, table string, attributeName string, autoIncrement bool, tableId, fieldId int, Driver model.Driver) pk {
	table = Driver.KeywordHandler(table)
	return pk{
		attributeStrings: createAttributeStrings(db, schema, table, attributeName, tableId, fieldId, Driver),
		autoIncrement:    autoIncrement}
}

type att struct {
	isDefault bool
	attributeStrings
}

func (a att) schema() *string {
	return a.schemaName
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

func createAtt(db *DB, attributeName string, schema *string, table string, tableId, fieldId int, isDefault bool, d model.Driver) att {
	return att{
		isDefault:        isDefault,
		attributeStrings: createAttributeStrings(db, schema, table, attributeName, tableId, fieldId, d)}
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
	b.query.ReturningID = &model.Attribute{Name: p.getAttributeName()}
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

func (p pk) getDefault() bool {
	return false
}

func (a att) getDefault() bool {
	return a.isDefault
}

func (m manyToOne) getDefault() bool {
	return m.isDefault
}

func (o oneToOne) getDefault() bool {
	return false
}

type aggregateResult struct {
	attributeName string
	tableName     string
	schemaName    *string
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

func (a aggregateResult) schema() *string {
	return a.schemaName
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
	schemaName    *string
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

func (f functionResult) schema() *string {
	return f.schemaName
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
