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

func (o *oneToOne) GetPrimaryKey() *pk {
	return o.pk
}

func (o *oneToOne) Table() []byte {
	return o.tableBytes
}

func createOneToOne(typeOf reflect.Type, targetTypeOf reflect.Type, Driver Driver, prefix string) *oneToOne {
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
		Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		Driver.KeywordHandler(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.tableBytes = []byte(Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())))
	mto.attributeName = Driver.KeywordHandler(utils.ColumnNamePattern(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.structAttributeName = prefix + typeOf.Name()
	return mto
}

type manyToOne struct {
	pk *pk
	attributeStrings
}

func (m *manyToOne) GetPrimaryKey() *pk {
	return m.pk
}

func (m *manyToOne) Table() []byte {
	return m.tableBytes
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, Driver Driver, prefix string) *manyToOne {
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
		Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())),
		Driver.KeywordHandler(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.tableBytes = []byte(Driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name())))
	mto.attributeName = Driver.KeywordHandler(utils.ColumnNamePattern(utils.ManyToOneNamePattern(prefix, typeOf.Name())))
	mto.structAttributeName = prefix + typeOf.Name()
	return mto
}

type attributeStrings struct {
	tableBytes          []byte
	selectName          string
	attributeName       string
	structAttributeName string
}

func createAttributeStrings(table []byte, attributeName string, Driver Driver) attributeStrings {
	return attributeStrings{
		tableBytes:          table,
		selectName:          fmt.Sprintf("%v.%v", string(table), Driver.KeywordHandler(utils.ColumnNamePattern(attributeName))),
		attributeName:       Driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		structAttributeName: attributeName,
	}
}

type pk struct {
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func (p *pk) GetPrimaryKey() *pk {
	return p
}

func (p *pk) Table() []byte {
	return p.tableBytes
}

func createPk(table []byte, attributeName string, autoIncrement bool, Driver Driver) *pk {
	//TODO:: Check this utils
	table = []byte(Driver.KeywordHandler(utils.TableNamePattern(string(table))))
	return &pk{
		attributeStrings: createAttributeStrings(table, attributeName, Driver),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	pk *pk
}

func (a *att) GetPrimaryKey() *pk {
	return a.pk
}

func (a *att) Table() []byte {
	return a.tableBytes
}

func createAtt(attributeName string, pk *pk, d Driver) *att {
	return &att{
		attributeStrings: createAttributeStrings(pk.tableBytes, attributeName, d), pk: pk}
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
		b.Sql.WriteString(p.attributeName)
		b.AttrNames = append(b.AttrNames, p.structAttributeName)
		return
	}
	b.Returning = b.Driver.Returning([]byte(p.attributeName))
	b.StructPkName = p.structAttributeName
}

func (a *att) BuildAttributeInsert(b *Builder) {
	b.Sql.WriteString(a.attributeName)
	b.AttrNames = append(b.AttrNames, a.structAttributeName)
}

func (m *manyToOne) BuildAttributeInsert(b *Builder) {
	b.Sql.WriteString(m.attributeName)
	b.AttrNames = append(b.AttrNames, m.structAttributeName)
}

func (o *oneToOne) BuildAttributeInsert(b *Builder) {
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
