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

	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if slices.Contains(fieldNames, field.Name) {
			table, prefix := checkTablePattern(tables, field)
			if table == "" && prefix == "" {
				continue
			}
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			err = handlerSliceMigrate(table, tables, field, valueOf.Field(i).Type().Elem(), valueOf, i, driver)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = handlerStructMigrate(table, field, valueOf.Field(i).Type(), valueOf, i, driver)
			if err != nil {
				return err
			}
		case reflect.Ptr:
			err = helperAttributeMigrate(table, tables, valueOf, field, i, true, driver)
			if err != nil {
				return err
			}
		default:
			err = helperAttributeMigrate(table, tables, valueOf, field, i, false, driver)
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

func handlerStructMigrate(table *TableMigrate, field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, driver Driver) error {
	switch targetTypeOf.Name() {
	case "Time":
		return migrateAtt(table, valueOf, field, i, driver)
	}
	return nil
}

func handlerSliceMigrate(table *TableMigrate, tables reflect.Value, field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, driver Driver) error {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		return helperAttributeMigrate(table, tables, valueOf, field, i, false, driver)
	}
	return nil
}

func isManyToOneMigrate(tables reflect.Value, typeOf reflect.Type, nullable bool, table, prefix, fieldName string, driver Driver) any {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createManyToOneMigrate(tables.Field(c).Elem().Type(), nullable, prefix, fieldName, driver)
					}
				}
			}
			if tableMtm := strings.ReplaceAll(typeOf.Name(), table, ""); tableMtm != typeOf.Name() {
				typeOfMtm := tables.FieldByName(tableMtm)
				if typeOfMtm.IsValid() && !typeOfMtm.IsZero() {
					typeOfMtm = typeOfMtm.Elem()
					for i := 0; i < typeOfMtm.NumField(); i++ {
						if typeOfMtm.Field(i).Kind() == reflect.Slice && typeOfMtm.Field(i).Type().Elem().Name() == table {
							return createManyToOneMigrate(tables.Field(c).Elem().Type(), nullable, prefix, fieldName, driver)
						}
					}
				}
			}
			return createOneToOneMigrate(tables.Field(c).Elem().Type(), nullable, prefix, fieldName, driver)
		}
	}
	return nil
}

func createManyToOneMigrate(typeOf reflect.Type, nullable bool, prefix, fieldName string, driver Driver) *ManyToOneMigrate {
	fieldPks := primaryKeys(typeOf)
	count := 0
	for i := range fieldPks {
		if fieldPks[i].Name == prefix {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto := new(ManyToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(prefix)
	mto.EscapingTargetTable = driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(fieldName)
	mto.EscapingName = driver.KeywordHandler(mto.Name)
	mto.Nullable = nullable
	return mto
}

func createOneToOneMigrate(typeOf reflect.Type, nullable bool, prefix, fieldName string, driver Driver) *OneToOneMigrate {
	fieldPks := primaryKeys(typeOf)
	count := 0
	for i := range fieldPks {
		if fieldPks[i].Name == prefix {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	mto := new(OneToOneMigrate)

	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(prefix)
	mto.EscapingTargetTable = driver.KeywordHandler(mto.TargetTable)
	mto.EscapingTargetColumn = driver.KeywordHandler(mto.TargetColumn)

	mto.Name = utils.ColumnNamePattern(fieldName)
	mto.EscapingName = driver.KeywordHandler(mto.Name)
	mto.Nullable = nullable
	return mto
}

func migratePk(typeOf reflect.Type, driver Driver) ([]*PrimaryKeyMigrate, []string, error) {
	var pks []*PrimaryKeyMigrate
	var fieldsNames []string

	id, valid := getId(typeOf)
	if valid {
		pks = make([]*PrimaryKeyMigrate, 1)
		fieldsNames = make([]string, 1)
		pks[0] = createMigratePk(id.Name, isAutoIncrement(id), getType(id), driver)
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
		pks[i] = createMigratePk(fields[i].Name, isAutoIncrement(fields[i]), getType(fields[i]), driver)
		fieldsNames[i] = fields[i].Name
	}
	return pks, fieldsNames, nil
}

func migrateAtt(table *TableMigrate, valueOf reflect.Value, field reflect.StructField, i int, driver Driver) error {
	at := createMigrateAtt(
		valueOf.Type().Field(i).Name,
		getType(field),
		field.Type.String()[0] == '*',
		driver,
	)
	table.Attributes = append(table.Attributes, *at)

	indexFunc := getIndex(field)
	if indexFunc != "" {
		for _, index := range strings.Split(indexFunc, ",") {
			indexName := getIndexValue(index, "n:")

			if indexName == "" {
				indexName = table.Name + "_idx_" + strings.ToLower(field.Name)
			}
			in := IndexMigrate{
				Name:         table.Name + "_" + indexName,
				EscapingName: driver.KeywordHandler(table.Name + "_" + indexName),
				Unique:       strings.Contains(index, "unique"),
				Attributes:   []AttributeMigrate{*at},
			}

			var i int
			if i = slices.IndexFunc(table.Indexes, func(i IndexMigrate) bool {
				return i.Name == in.Name && i.Unique == in.Unique
			}); i == -1 {
				if c := slices.IndexFunc(table.Indexes, func(i IndexMigrate) bool {
					return i.Name == in.Name && i.Unique != in.Unique
				}); c != -1 {
					return fmt.Errorf(`goe: struct "%v" have two or more indexes with same name but different uniqueness "%v"`, table.Name, in.Name)
				}

				table.Indexes = append(table.Indexes, in)
				continue
			}
			table.Indexes[i].Attributes = append(table.Indexes[i].Attributes, *at)
		}
	}

	tagValue := field.Tag.Get("goe")
	if tagValueExist(tagValue, "unique") {
		in := IndexMigrate{
			Name:         table.Name + "_idx_" + strings.ToLower(field.Name),
			EscapingName: driver.KeywordHandler(table.Name + "_idx_" + strings.ToLower(field.Name)),
			Unique:       true,
			Attributes:   []AttributeMigrate{*at},
		}
		table.Indexes = append(table.Indexes, in)
	}

	if tagValueExist(tagValue, "index") {
		in := IndexMigrate{
			Name:         table.Name + "_idx_" + strings.ToLower(field.Name),
			EscapingName: driver.KeywordHandler(table.Name + "_idx_" + strings.ToLower(field.Name)),
			Unique:       false,
			Attributes:   []AttributeMigrate{*at},
		}
		table.Indexes = append(table.Indexes, in)
	}
	return nil
}

func getType(field reflect.StructField) string {
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

func helperAttributeMigrate(tbl *TableMigrate, tables reflect.Value, valueOf reflect.Value, field reflect.StructField, i int, nullable bool, driver Driver) error {
	table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
	if table != "" {
		if mto := isManyToOneMigrate(tables, valueOf.Type(), nullable, table, prefix, valueOf.Type().Field(i).Name, driver); mto != nil {
			switch v := mto.(type) {
			case *ManyToOneMigrate:
				if v == nil {
					return migrateAtt(tbl, valueOf, field, i, driver)
				}
				v.DataType = getType(field)
				tbl.ManyToOnes = append(tbl.ManyToOnes, *v)
			case *OneToOneMigrate:
				if v == nil {
					return migrateAtt(tbl, valueOf, field, i, driver)
				}
				v.DataType = getType(field)
				tbl.OneToOnes = append(tbl.OneToOnes, *v)
			}
			return nil
		}
	}
	return migrateAtt(tbl, valueOf, field, i, driver)
}
