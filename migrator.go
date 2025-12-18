package goe

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/utils"
)

func migrateFrom(db any, driver model.Driver) *model.Migrator {
	valueOf := reflect.ValueOf(db).Elem()

	schemasMap := make(map[string]*string)
	for i := range valueOf.NumField() - 1 {
		if strings.Contains(valueOf.Type().Field(i).Tag.Get("goe"), "schema") || strings.HasSuffix(valueOf.Field(i).Elem().Type().Name(), "Schema") {
			schema := driver.KeywordHandler(utils.ColumnNamePattern(valueOf.Field(i).Elem().Type().Name()))
			for f := range valueOf.Field(i).Elem().NumField() {
				schemasMap[valueOf.Field(i).Elem().Field(f).Elem().Type().Name()] = &schema
			}
		}
	}

	migrator := new(model.Migrator)
	migrator.Tables = make(map[string]*model.TableMigrate)
	for i := range valueOf.NumField() - 1 {
		if strings.Contains(valueOf.Type().Field(i).Tag.Get("goe"), "schema") || strings.HasSuffix(valueOf.Field(i).Elem().Type().Name(), "Schema") {
			schema := driver.KeywordHandler(utils.ColumnNamePattern(valueOf.Field(i).Elem().Type().Name()))
			migrator.Schemas = append(migrator.Schemas, schema)
			for f := range valueOf.Field(i).Elem().NumField() {
				migrator.Error = typeField(valueOf, valueOf.Field(i).Elem().Field(f), migrator, driver, &schema, schemasMap)
				if migrator.Error != nil {
					return migrator
				}
			}
			continue
		}

		migrator.Error = typeField(valueOf, valueOf.Field(i), migrator, driver, nil, schemasMap)
		if migrator.Error != nil {
			return migrator
		}
	}

	return migrator
}

func typeField(tables reflect.Value, valueOf reflect.Value, migrator *model.Migrator, driver model.Driver, schema *string, schemasMap map[string]*string) error {
	valueOf = valueOf.Elem()
	pks, fieldNames, err := migratePk(valueOf.Type(), driver)
	if err != nil {
		return err
	}
	table := new(model.TableMigrate)

	table.Name = utils.TableNamePattern(valueOf.Type().Name())
	table.Schema = schema
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
				schemasMap: schemasMap,
			}, helperAttributeMigrate)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = handlerStruct(body{
				fieldId:     fieldId,
				driver:      driver,
				nullable:    isNullable(field),
				fieldTypeOf: valueOf.Field(fieldId).Type(),
				valueOf:     valueOf,
				migrate: &infosMigrate{
					table: table,
					field: field,
				},
				schemasMap: schemasMap,
			}, migrateAtt)
			if err != nil {
				return err
			}
		case reflect.Pointer:
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
				schemasMap: schemasMap,
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
				schemasMap: schemasMap,
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

	mto := new(model.ManyToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(b.prefixName)
	mto.TargetSchema = b.schemasMap[typeOf.Name()]
	mto.EscapingTargetTable = b.driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = b.driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(b.fieldName)
	mto.EscapingName = b.driver.KeywordHandler(mto.Name)
	mto.Nullable = b.nullable
	mto.Default = getTagValue(b.migrate.field.Tag.Get("goe"), "default:")
	if err := checkIndex(b, mto.AttributeMigrate, true); err != nil {
		panic(err)
	}
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

	mto := new(model.OneToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(b.prefixName)
	mto.TargetSchema = b.schemasMap[typeOf.Name()]
	mto.EscapingTargetTable = b.driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = b.driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(b.fieldName)
	mto.EscapingName = b.driver.KeywordHandler(mto.Name)
	mto.Nullable = b.nullable
	if err := checkIndex(b, mto.AttributeMigrate, true); err != nil {
		panic(err)
	}
	return mto
}

func migratePk(typeOf reflect.Type, driver model.Driver) ([]*model.PrimaryKeyMigrate, []string, error) {
	fields := getPks(typeOf)
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("goe: struct %q don't have a primary key setted", typeOf.Name())
	}

	pks := make([]*model.PrimaryKeyMigrate, len(fields))
	fieldsNames := make([]string, len(fields))
	for i := range fields {
		pks[i] = createMigratePk(fields[i].Name, isAutoIncrement(fields[i]), getTagType(fields[i]), getTagValue(fields[i].Tag.Get("goe"), "default:"), driver)
		fieldsNames[i] = fields[i].Name
	}
	return pks, fieldsNames, nil
}

func isAutoIncrement(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int")
}

func migrateAtt(b body) error {
	at := createMigrateAtt(
		b.migrate.field.Name,
		getTagType(b.migrate.field),
		b.nullable,
		getTagValue(b.migrate.field.Tag.Get("goe"), "default:"),
		b.driver,
	)
	b.migrate.table.Attributes = append(b.migrate.table.Attributes, at)

	return checkIndex(b, at, false)
}

func getTagType(field reflect.StructField) string {
	value := getTagValue(field.Tag.Get("goe"), "type:")
	if value != "" {
		return strings.ReplaceAll(value, " ", "")
	}
	dataType := field.Type.String()
	if dataType[0] == '*' {
		return dataType[1:]
	}
	return dataType
}

func isNullable(field reflect.StructField) bool {
	dataType := field.Type.String()
	return strings.HasPrefix(dataType, "sql.Null")
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

func createMigratePk(attributeName string, autoIncrement bool, dataType, defaultTag string, driver model.Driver) *model.PrimaryKeyMigrate {
	return &model.PrimaryKeyMigrate{
		AttributeMigrate: model.AttributeMigrate{
			Name:         utils.ColumnNamePattern(attributeName),
			EscapingName: driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
			DataType:     dataType,
			Default:      defaultTag,
		},
		AutoIncrement: autoIncrement,
	}
}

func createMigrateAtt(attributeName string, dataType string, nullable bool, defaultValue string, driver model.Driver) model.AttributeMigrate {
	return model.AttributeMigrate{
		Name:         utils.ColumnNamePattern(attributeName),
		EscapingName: driver.KeywordHandler(utils.ColumnNamePattern(attributeName)),
		DataType:     dataType,
		Nullable:     nullable,
		Default:      defaultValue,
	}
}

func helperAttributeMigrate(b body) error {
	table, prefix := foreignKeyNamePattern(b.tables, b.migrate.field.Name)
	if table != "" {
		b.stringInfos = stringInfos{prefixName: prefix, tableName: table, fieldName: b.migrate.field.Name}
		if mto := isManyToOne(b, createManyToOneMigrate, createOneToOneMigrate); mto != nil {
			switch v := mto.(type) {
			case *model.ManyToOneMigrate:
				if v == nil {
					return migrateAtt(b)
				}
				v.DataType = getTagType(b.migrate.field)
				b.migrate.table.ManyToOnes = append(b.migrate.table.ManyToOnes, *v)
			case *model.OneToOneMigrate:
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

func checkIndex(b body, at model.AttributeMigrate, skipUnique bool) error {
	indexFunc := getIndex(b.migrate.field)
	if indexFunc != "" {
		for _, index := range strings.Split(indexFunc, ",") {
			indexName := getIndexValue(index, "n:")

			if indexName == "" {
				indexName = b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)
			}
			in := model.IndexMigrate{
				Name:         b.migrate.table.Name + "_" + indexName,
				EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_" + indexName),
				Unique:       strings.Contains(index, "unique"),
				Func:         strings.ToLower(getIndexValue(index, "f:")),
				Attributes:   []model.AttributeMigrate{at},
			}

			var i int
			if i = slices.IndexFunc(b.migrate.table.Indexes, func(i model.IndexMigrate) bool {
				return i.Name == in.Name && i.Unique == in.Unique && i.Func == in.Func
			}); i == -1 {
				if c := slices.IndexFunc(b.migrate.table.Indexes, func(i model.IndexMigrate) bool {
					return i.Name == in.Name && (i.Unique != in.Unique || i.Func != in.Func)
				}); c != -1 {
					return fmt.Errorf(`goe: struct "%v" have two or more indexes with same name but different uniqueness/function "%v"`, b.migrate.table.Name, in.Name)
				}

				b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
				continue
			}
			b.migrate.table.Indexes[i].Attributes = append(b.migrate.table.Indexes[i].Attributes, at)
		}
	}

	tagValue := b.migrate.field.Tag.Get("goe")
	if !skipUnique && tagValueExist(tagValue, "unique") {
		in := model.IndexMigrate{
			Name:         b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name),
			EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)),
			Unique:       true,
			Attributes:   []model.AttributeMigrate{at},
		}
		b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
	}

	if tagValueExist(tagValue, "index") {
		in := model.IndexMigrate{
			Name:         b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name),
			EscapingName: b.driver.KeywordHandler(b.migrate.table.Name + "_idx_" + strings.ToLower(b.migrate.field.Name)),
			Unique:       false,
			Attributes:   []model.AttributeMigrate{at},
		}
		b.migrate.table.Indexes = append(b.migrate.table.Indexes, in)
	}
	return nil
}
