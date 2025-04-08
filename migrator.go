package goe

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/olauro/goe/utils"
)

type Migrator struct {
	Tables map[string]*TableMigrate
	Error  error
}

type TableMigrate struct {
	Name         string
	EscapingName string
	Migrated     bool
	PrimaryKeys  []PrimaryKeyMigrate
	Attributes   []AttributeMigrate
	ManyToOnes   []ManyToOneMigrate
	OneToOnes    []OneToOneMigrate
	Indexes      []IndexMigrate
}

type IndexMigrate struct {
	Name         string
	EscapingName string
	Unique       bool
	Attributes   []AttributeMigrate
}

type PrimaryKeyMigrate struct {
	AutoIncrement bool
	Name          string
	EscapingName  string
	DataType      string
}

type AttributeMigrate struct {
	Nullable     bool
	Name         string
	EscapingName string
	DataType     string
}

type OneToOneMigrate struct {
	AttributeMigrate
	TargetTable          string
	TargetColumn         string
	EscapingTargetTable  string
	EscapingTargetColumn string
}

type ManyToOneMigrate struct {
	AttributeMigrate
	TargetTable          string
	TargetColumn         string
	EscapingTargetTable  string
	EscapingTargetColumn string
}

func migrateFrom(db any, driver Driver) *Migrator {
	valueOf := reflect.ValueOf(db).Elem()

	migrator := new(Migrator)
	migrator.Tables = make(map[string]*TableMigrate)
	for i := range valueOf.NumField() - 1 {
		migrator.Error = typeField(valueOf, valueOf.Field(i).Elem(), migrator, driver)
		if migrator.Error != nil {
			return migrator
		}
	}

	return migrator
}

func typeField(tables reflect.Value, valueOf reflect.Value, migrator *Migrator, driver Driver) error {
	pks, fieldNames, err := migratePk(valueOf.Type(), driver)
	if err != nil {
		return err
	}
	table := new(TableMigrate)

	table.Name = utils.TableNamePattern(valueOf.Type().Name())
	var field reflect.StructField

	for fieldId := range valueOf.NumField() {
		field = valueOf.Type().Field(fieldId)
		if skipPrimaryKey(fieldNames, field.Name, tables, field) {
			continue
		}
		switch valueOf.Field(fieldId).Kind() {
		case reflect.Slice:
			err = handlerSlice(body{
				fieldId:     fieldId,
				driver:      driver,
				tables:      tables,
				fieldTypeOf: valueOf.Field(fieldId).Type().Elem(),
				typeOf:      valueOf.Type(),
				valueOf:     valueOf,
				migrate: &infosMigrate{
					table:      table,
					field:      field,
					fieldNames: fieldNames,
				},
			}, helperAttributeMigrate)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = handlerStruct(body{
				fieldId:     fieldId,
				driver:      driver,
				fieldTypeOf: valueOf.Field(fieldId).Type(),
				valueOf:     valueOf,
				migrate: &infosMigrate{
					table: table,
					field: field,
				},
			}, migrateAtt)
			if err != nil {
				return err
			}
		case reflect.Ptr:
			err = helperAttributeMigrate(body{
				fieldId:  fieldId,
				driver:   driver,
				nullable: true,
				tables:   tables,
				valueOf:  valueOf,
				typeOf:   valueOf.Type(),
				migrate: &infosMigrate{
					table:      table,
					field:      field,
					fieldNames: fieldNames,
				},
			})
			if err != nil {
				return err
			}
		default:
			err = helperAttributeMigrate(body{
				fieldId: fieldId,
				driver:  driver,
				tables:  tables,
				valueOf: valueOf,
				typeOf:  valueOf.Type(),
				migrate: &infosMigrate{
					table:      table,
					field:      field,
					fieldNames: fieldNames,
				},
			})
			if err != nil {
				return err
			}
		}
	}

	for _, pk := range pks {
		table.PrimaryKeys = append(table.PrimaryKeys, *pk)
	}

	table.EscapingName = driver.KeywordHandler(table.Name)
	migrator.Tables[table.Name] = table
	return nil
}

func createManyToOneMigrate(b body, typeOf reflect.Type) any {
	fieldPks := primaryKeys(typeOf)
	count := 0
	for i := range fieldPks {
		if fieldPks[i].Name == b.prefixName {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto := new(ManyToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(b.prefixName)
	mto.EscapingTargetTable = b.driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = b.driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(b.fieldName)
	mto.EscapingName = b.driver.KeywordHandler(mto.Name)
	mto.Nullable = b.nullable
	return mto
}

func createOneToOneMigrate(b body, typeOf reflect.Type) any {
	fieldPks := primaryKeys(typeOf)
	count := 0
	for i := range fieldPks {
		if fieldPks[i].Name == b.prefixName {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto := new(OneToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(b.prefixName)
	mto.EscapingTargetTable = b.driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = b.driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(b.fieldName)
	mto.EscapingName = b.driver.KeywordHandler(mto.Name)
	mto.Nullable = b.nullable
	return mto
}

func migratePk(typeOf reflect.Type, driver Driver) ([]*PrimaryKeyMigrate, []string, error) {
	var pks []*PrimaryKeyMigrate
	var fieldsNames []string

	id, valid := getId(typeOf)
	if valid {
		pks = make([]*PrimaryKeyMigrate, 1)
		fieldsNames = make([]string, 1)
		pks[0] = createMigratePk(id.Name, isAutoIncrement(id), getTagType(id), driver)
		fieldsNames[0] = id.Name
		return pks, fieldsNames, nil
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("goe: struct %q don't have a primary key setted", typeOf.Name())
	}

	pks = make([]*PrimaryKeyMigrate, len(fields))
	fieldsNames = make([]string, len(fields))
	for i := range fields {
		pks[i] = createMigratePk(fields[i].Name, isAutoIncrement(fields[i]), getTagType(fields[i]), driver)
		fieldsNames[i] = fields[i].Name
	}
	return pks, fieldsNames, nil
}

func migrateAtt(b body) error {
	at := createMigrateAtt(
		b.migrate.field.Name,
		getTagType(b.migrate.field),
		b.nullable,
		b.driver,
	)
	b.migrate.table.Attributes = append(b.migrate.table.Attributes, *at)

	indexFunc := getIndex(b.migrate.field)
	if indexFunc != "" {
		for _, index := range strings.Split(indexFunc, ",") {
			indexName := getIndexValue(index, "n:")

			if indexName == "" {
				indexName = b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)
			}
			in := IndexMigrate{
				Name:         b.migrate.table.Name + "_" + indexName,
				EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_" + indexName),
				Unique:       strings.Contains(index, "unique"),
				Attributes:   []AttributeMigrate{*at},
			}

			var i int
			if i = slices.IndexFunc(b.migrate.table.Indexes, func(i IndexMigrate) bool {
				return i.Name == in.Name && i.Unique == in.Unique
			}); i == -1 {
				if c := slices.IndexFunc(b.migrate.table.Indexes, func(i IndexMigrate) bool {
					return i.Name == in.Name && i.Unique != in.Unique
				}); c != -1 {
					return fmt.Errorf(`goe: struct "%v" have two or more indexes with same name but different uniqueness "%v"`, b.migrate.table.Name, in.Name)
				}

				b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
				continue
			}
			b.migrate.table.Indexes[i].Attributes = append(b.migrate.table.Indexes[i].Attributes, *at)
		}
	}

	tagValue := b.migrate.field.Tag.Get("goe")
	if tagValueExist(tagValue, "unique") {
		in := IndexMigrate{
			Name:         b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name),
			EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)),
			Unique:       true,
			Attributes:   []AttributeMigrate{*at},
		}
		b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
	}

	if tagValueExist(tagValue, "index") {
		in := IndexMigrate{
			Name:         b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name),
			EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)),
			Unique:       false,
			Attributes:   []AttributeMigrate{*at},
		}
		b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
	}
	return nil
}

func getTagType(field reflect.StructField) string {
	value := getTagValue(field.Tag.Get("goe"), "type:")
	if value != "" {
		return value
	}
	dataType := field.Type.String()
	if dataType[0] == '*' {
		return dataType[1:]
	}
	return dataType
}

func getIndex(field reflect.StructField) string {
	value := getTagValue(field.Tag.Get("goe"), "index(")
	if value != "" {
		return value[0 : len(value)-1]
	}
	return ""
}

func tagValueExist(tag string, subTag string) bool {
	values := strings.Split(tag, ";")
	for _, v := range values {
		if v == subTag {
			return true
		}
	}
	return false
}

func getIndexValue(valueTag string, tag string) string {
	values := strings.Split(valueTag, " ")
	for _, v := range values {
		if _, value, ok := strings.Cut(v, tag); ok {
			return value
		}
	}
	return ""
}

func createMigratePk(attributeName string, autoIncrement bool, dataType string, driver Driver) *PrimaryKeyMigrate {
	return &PrimaryKeyMigrate{
		Name:          utils.ColumnNamePattern(attributeName),
		EscapingName:  driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		DataType:      dataType,
		AutoIncrement: autoIncrement}
}

func createMigrateAtt(attributeName string, dataType string, nullable bool, driver Driver) *AttributeMigrate {
	return &AttributeMigrate{
		Name:         utils.ColumnNamePattern(attributeName),
		EscapingName: driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		DataType:     dataType,
		Nullable:     nullable,
	}
}

func helperAttributeMigrate(b body) error {
	table, prefix := checkTablePattern(b.tables, b.migrate.field)
	if table != "" {
		b.stringInfos = stringInfos{prefixName: prefix, tableName: table, fieldName: b.migrate.field.Name}
		if mto := isManyToOne(b, createManyToOneMigrate, createOneToOneMigrate); mto != nil {
			switch v := mto.(type) {
			case *ManyToOneMigrate:
				if v == nil {
					return migrateAtt(b)
				}
				v.DataType = getTagType(b.migrate.field)
				b.migrate.table.ManyToOnes = append(b.migrate.table.ManyToOnes, *v)
			case *OneToOneMigrate:
				if v == nil {
					if slices.Contains(b.migrate.fieldNames, b.migrate.field.Name) {
						return nil
					}
					return migrateAtt(b)
				}
				v.DataType = getTagType(b.migrate.field)
				b.migrate.table.OneToOnes = append(b.migrate.table.OneToOnes, *v)
			}
			return nil
		}
	}
	return migrateAtt(b)
}
