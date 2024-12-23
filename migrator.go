package goe

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/olauro/goe/utils"
)

type Migrator struct {
	Tables []any
	Error  error
}

func MigrateFrom(db any) *Migrator {
	valueOf := reflect.ValueOf(db).Elem()

	migrator := new(Migrator)
	migrator.Tables = make([]any, 0)
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Type().Elem().Name() != "DB" {
			migrator.Error = typeField(valueOf, valueOf.Field(i).Elem(), migrator)
			if migrator.Error != nil {
				return migrator
			}
		}
	}

	return migrator
}

func typeField(tables reflect.Value, valueOf reflect.Value, migrator *Migrator) error {
	pks, fieldNames, err := migratePk(valueOf.Type())
	if err != nil {
		return err
	}

	for _, pk := range pks {
		migrator.Tables = append(migrator.Tables, pk)
	}
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if slices.Contains(fieldNames, field.Name) {
			//TODO: Check this
			table, prefix := checkTablePattern(tables, field)
			if table == "" && prefix == "" {
				continue
			}
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			err = handlerSliceMigrate(tables, field, valueOf.Field(i).Type().Elem(), valueOf, i, pks, migrator)
			if err != nil {
				return err
			}
		case reflect.Struct:
			handlerStructMigrate(field, valueOf.Field(i).Type(), valueOf, i, pks[0], migrator)
		case reflect.Ptr:
			table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
			if table != "" {
				if mto := isMigrateManyToOne(tables, valueOf.Type(), true, table, prefix); mto != nil {
					switch v := mto.(type) {
					case *MigrateManyToOne:
						if v == nil {
							migrateAtt(valueOf, field, i, pks[0], migrator)
							continue
						}
					case *MigrateOneToOne:
						if v == nil {
							migrateAtt(valueOf, field, i, pks[0], migrator)
							continue
						}
					}

					key := utils.TableNamePattern(table)
					for _, pk := range pks {
						if pk.AttributeName == strings.ToLower(prefix) || pk.AttributeName == strings.ToLower(prefix+table) {
							pk.Fks[key] = mto
						}
					}
					continue
				}
			}
			migrateAtt(valueOf, field, i, pks[0], migrator)
		default:
			table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
			if table != "" {
				if mto := isMigrateManyToOne(tables, valueOf.Type(), false, table, prefix); mto != nil {
					switch v := mto.(type) {
					case *MigrateManyToOne:
						if v == nil {
							migrateAtt(valueOf, field, i, pks[0], migrator)
							continue
						}
					case *MigrateOneToOne:
						if v == nil {
							migrateAtt(valueOf, field, i, pks[0], migrator)
							continue
						}
					}
					key := utils.TableNamePattern(table)
					for _, pk := range pks {
						if pk.AttributeName == strings.ToLower(prefix) || pk.AttributeName == strings.ToLower(prefix+table) {
							pk.Fks[key] = mto
						}
					}
					continue
				}
			}
			migrateAtt(valueOf, field, i, pks[0], migrator)
		}
	}
	return nil
}

func handlerStructMigrate(field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *MigratePk, migrator *Migrator) {
	switch targetTypeOf.Name() {
	case "Time":
		migrateAtt(valueOf, field, i, p, migrator)
	}
}

func handlerSliceMigrate(tables reflect.Value, field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, pks []*MigratePk, migrator *Migrator) error {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
		if table != "" {
			if mto := isMigrateManyToOne(tables, valueOf.Type(), false, table, prefix); mto != nil {
				switch v := mto.(type) {
				case *MigrateManyToOne:
					if v == nil {
						migrateAtt(valueOf, field, i, pks[0], migrator)
						return nil
					}
				case *MigrateOneToOne:
					if v == nil {
						migrateAtt(valueOf, field, i, pks[0], migrator)
						return nil
					}
				}
				key := utils.TableNamePattern(table)
				for _, pk := range pks {
					if pk.AttributeName == strings.ToLower(prefix) || pk.AttributeName == strings.ToLower(prefix+table) {
						pk.Fks[key] = mto
					}
				}
				return nil
			}
		}
		migrateAtt(valueOf, field, i, pks[0], migrator)
	}
	return nil
}

func isMigrateManyToOne(tables reflect.Value, typeOf reflect.Type, nullable bool, table, prefix string) any {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createMigrateManyToOne(tables.Field(c).Elem().Type(), typeOf, nullable, prefix)
					}
				}
			}
			if tableMtm := strings.ReplaceAll(typeOf.Name(), table, ""); tableMtm != typeOf.Name() {
				typeOfMtm := tables.FieldByName(tableMtm)
				if typeOfMtm.IsValid() && !typeOfMtm.IsZero() {
					typeOfMtm = typeOfMtm.Elem()
					for i := 0; i < typeOfMtm.NumField(); i++ {
						if typeOfMtm.Field(i).Kind() == reflect.Slice && typeOfMtm.Field(i).Type().Elem().Name() == table {
							return createMigrateManyToOne(tables.Field(c).Elem().Type(), typeOf, nullable, prefix)
						}
					}
				}
			}
			return createMigrateOneToOne(tables.Field(c).Elem().Type(), typeOf, nullable, prefix)
		}
	}
	return nil
}

func createMigrateManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, nullable bool, prefix string) *MigrateManyToOne {
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

	mto := new(MigrateManyToOne)
	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(prefix)
	mto.Id = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(prefix, typeOf.Name()))
	mto.Nullable = nullable
	return mto
}

func createMigrateOneToOne(typeOf reflect.Type, targetTypeOf reflect.Type, nullable bool, prefix string) *MigrateOneToOne {
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

	mto := new(MigrateOneToOne)
	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.TargetColumn = utils.ColumnNamePattern(prefix)
	mto.Id = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(prefix, typeOf.Name()))
	mto.Nullable = nullable
	return mto
}

type MigratePk struct {
	Table         string
	AutoIncrement bool
	Fks           map[string]any
	AttributeName string
	DataType      string
}

type MigrateAtt struct {
	Nullable      bool
	Index         string
	Pk            *MigratePk
	AttributeName string
	DataType      string
}

type MigrateOneToOne struct {
	TargetTable  string
	TargetColumn string
	Nullable     bool
	Id           string
}

type MigrateManyToOne struct {
	TargetTable  string
	TargetColumn string
	Nullable     bool
	Id           string
}

type AttributeStrings struct {
	AttributeName string
	DataType      string
}

func migratePk(typeOf reflect.Type) ([]*MigratePk, []string, error) {
	var pks []*MigratePk
	var fieldsNames []string

	id, valid := typeOf.FieldByName("Id")
	if valid {
		pks = make([]*MigratePk, 1)
		fieldsNames = make([]string, 1)
		pks[0] = createMigratePk(typeOf.Name(), id.Name, isAutoIncrement(id), getType(id))
		fieldsNames[0] = id.Name
		return pks, fieldsNames, nil
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("%w: struct %q don't have a primary key setted", ErrStructWithoutPrimaryKey, typeOf.Name())
	}

	pks = make([]*MigratePk, len(fields))
	fieldsNames = make([]string, len(fields))
	for i := range fields {
		pks[i] = createMigratePk(typeOf.Name(), fields[i].Name, isAutoIncrement(fields[i]), getType(fields[i]))
		fieldsNames[i] = fields[i].Name
	}
	return pks, fieldsNames, nil
}

func migrateAtt(valueOf reflect.Value, field reflect.StructField, i int, pk *MigratePk, m *Migrator) {
	at := createMigrateAtt(
		valueOf.Type().Field(i).Name,
		pk,
		getType(field),
		field.Type.String()[0] == '*',
		getIndex(field),
	)
	m.Tables = append(m.Tables, at)
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

func createMigratePk(table string, attributeName string, autoIncrement bool, dataType string) *MigratePk {
	return &MigratePk{
		Table:         utils.TableNamePattern(table),
		AttributeName: utils.ColumnNamePattern(attributeName),
		DataType:      dataType,
		AutoIncrement: autoIncrement,
		Fks:           make(map[string]any)}
}

func createMigrateAtt(attributeName string, pk *MigratePk, dataType string, nullable bool, index string) *MigrateAtt {
	return &MigrateAtt{
		AttributeName: utils.ColumnNamePattern(attributeName),
		DataType:      dataType,
		Pk:            pk,
		Nullable:      nullable,
		Index:         index,
	}
}
