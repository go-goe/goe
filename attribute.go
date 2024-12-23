package goe

import (
	"fmt"
	"reflect"

	"github.com/olauro/goe/utils"
)

type oneToOne struct {
	pk *pk
	attributeStrings
}

func (o *oneToOne) getPrimaryKey() *pk {
	return o.pk
}

func (o *oneToOne) table() []byte {
	return o.tableBytes
}

func createOneToOne(typeOf reflect.Type, targetTypeOf reflect.Type, driver Driver, prefix string) *oneToOne {
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

	mto.selectName = fmt.Sprintf("%v.%v",
		driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		driver.KeywordHandler(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.tableBytes = []byte(driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())))
	mto.attributeName = driver.KeywordHandler(utils.ColumnNamePattern(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.structAttributeName = prefix + typeOf.Name()
	return mto
}

type manyToOne struct {
	pk *pk
	attributeStrings
}

func (m *manyToOne) getPrimaryKey() *pk {
	return m.pk
}

func (m *manyToOne) table() []byte {
	return m.tableBytes
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, driver Driver, prefix string) *manyToOne {
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

	mto.selectName = fmt.Sprintf("%v.%v",
		driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		driver.KeywordHandler(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.tableBytes = []byte(driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())))
	mto.attributeName = driver.KeywordHandler(utils.ColumnNamePattern(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.structAttributeName = prefix + typeOf.Name()
	return mto
}

type attributeStrings struct {
	tableBytes          []byte
	selectName          string
	attributeName       string
	structAttributeName string
}

func createAttributeStrings(table []byte, attributeName string, driver Driver) attributeStrings {
	return attributeStrings{
		tableBytes:          table,
		selectName:          fmt.Sprintf("%v.%v", string(table), driver.KeywordHandler(utils.ColumnNamePattern(attributeName))),
		attributeName:       driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		structAttributeName: attributeName,
	}
}

type pk struct {
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func (p *pk) getPrimaryKey() *pk {
	return p
}

func (p *pk) table() []byte {
	return p.tableBytes
}

func createPk(table []byte, attributeName string, autoIncrement bool, driver Driver) *pk {
	//TODO:: Check this utils
	table = []byte(driver.KeywordHandler(utils.TableNamePattern(string(table))))
	return &pk{
		attributeStrings: createAttributeStrings(table, attributeName, driver),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	pk *pk
}

func (a *att) getPrimaryKey() *pk {
	return a.pk
}

func (a *att) table() []byte {
	return a.tableBytes
}

func createAtt(attributeName string, pk *pk, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(pk.tableBytes, attributeName, d), pk: pk}
}

func (p *pk) buildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(p.selectName)
	b.structColumns[i] = p.structAttributeName
}

func (a *att) buildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(a.selectName)
	b.structColumns[i] = a.structAttributeName
}

func (m *manyToOne) buildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(m.selectName)
	b.structColumns[i] = m.structAttributeName
}

func (o *oneToOne) buildAttributeSelect(b *builder, i int) {
	b.sql.WriteString(o.selectName)
	b.structColumns[i] = o.structAttributeName
}

func (p *pk) buildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.sql.WriteString(p.attributeName)
		b.attrNames = append(b.attrNames, p.structAttributeName)
		return
	}
	b.returning = b.driver.Returning([]byte(p.attributeName))
	b.structPkName = p.structAttributeName
}

func (a *att) buildAttributeInsert(b *builder) {
	b.sql.WriteString(a.attributeName)
	b.attrNames = append(b.attrNames, a.structAttributeName)
}

func (m *manyToOne) buildAttributeInsert(b *builder) {
	b.sql.WriteString(m.attributeName)
	b.attrNames = append(b.attrNames, m.structAttributeName)
}

func (o *oneToOne) buildAttributeInsert(b *builder) {
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
	function string
	field    field
}

func createAggregate(function string, f field) aggregate {
	return aggregate{function: function, field: f}
}

func (a aggregate) String() string {
	return fmt.Sprintf("%v(%v)", a.function, a.field.getSelect())
}
