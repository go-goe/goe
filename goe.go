package goe

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/utils"
)

func init() {
	addrMap = &goeMap{mapField: make(map[uintptr]field)}
}

// Open opens a database connection
//
// # Example
//
//	goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.Config{}))
func Open[T any](driver model.Driver) (*T, error) {
	driver.GetDatabaseConfig().Init(driver.Name(), driver.ErrorTranslator())

	err := driver.Init()
	if err != nil {
		return nil, driver.GetDatabaseConfig().ErrorHandler(context.TODO(), err)
	}

	db := new(T)
	valueOf := reflect.ValueOf(db).Elem()
	if valueOf.Kind() != reflect.Struct {
		return nil, errors.New("goe: invalid database, the target needs to be a struct")
	}

	dbId := valueOf.NumField() - 1
	if valueOf.Field(dbId).Type().Elem().Name() != "DB" {
		return nil, errors.New("goe: invalid database, last struct field needs to be goe.DB")
	}

	dbTarget := new(DB)
	valueOf.Field(dbId).Set(reflect.ValueOf(dbTarget))

	// set value for Fields
	for i := range dbId {
		field := valueOf.Field(i)
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
			if utils.IsFieldHasSchema(valueOf, i) {
				for f := range field.Elem().NumField() {
					elem := field.Elem().Field(f)
					elem.Set(reflect.New(elem.Type().Elem()))
				}
				continue
			}
		}
	}

	var schemas []string
	tableId := 0
	// init Fields
	for f := range dbId {
		elem := valueOf.Field(f).Elem()
		if utils.IsFieldHasSchema(valueOf, f) {
			schema := driver.KeywordHandler(utils.ColumnNamePattern(elem.Type().Name()))
			schemas = append(schemas, schema)
			for i := range elem.NumField() {
				tableId += i + 1
				err = initField(&schema, valueOf, elem.Field(i).Elem(), dbTarget, tableId, driver)
				if err != nil {
					return nil, err
				}
			}
			continue
		}
		tableId++
		err = initField(nil, valueOf, elem, dbTarget, tableId, driver)
		if err != nil {
			return nil, err
		}
	}

	dc := driver.GetDatabaseConfig()
	dc.SetSchemas(schemas)
	if ic := dc.InitCallback(); ic != nil {
		if err = ic(); err != nil {
			return nil, err
		}
	}

	dbTarget.driver = driver
	return db, nil
}

// data used for map
type infosMap struct {
	db      *DB
	pks     []pk
	tableId int
	addr    uintptr
}

// data used for migrate
type infosMigrate struct {
	field      reflect.StructField
	table      *model.TableMigrate
	fieldNames []string
}

type stringInfos struct {
	prefixName string
	tableName  string
	fieldName  string
}

type body struct {
	tables      reflect.Value // database value of
	valueOf     reflect.Value // struct value of
	typeOf      reflect.Type  // struct type of
	fieldTypeOf reflect.Type
	mapp        *infosMap     // used on map
	migrate     *infosMigrate // used on migrate
	schemasMap  map[string]*string
	fieldId     int
	driver      model.Driver
	nullable    bool
	schema      *string
	stringInfos
}

func skipPrimaryKey[T comparable](slice []T, value T, tables reflect.Value, field reflect.StructField) bool {
	if slices.Contains(slice, value) {
		table, prefix := foreignKeyNamePattern(tables, field.Name)
		if table == "" && prefix == "" {
			return true
		}
	}
	return false
}

func initField(schema *string, tables reflect.Value, valueOf reflect.Value, db *DB, tableId int, driver model.Driver) error {
	pks, fieldIds, err := getPk(db, schema, valueOf, tableId, driver)
	if err != nil {
		return err
	}

	for fieldId := range valueOf.NumField() {
		field := valueOf.Type().Field(fieldId)
		if skipPrimaryKey(fieldIds, fieldId, tables, field) {
			continue
		}
		addr := uintptr(valueOf.Field(fieldId).Addr().UnsafePointer())
		switch valueOf.Field(fieldId).Kind() {
		case reflect.Slice:
			err = handlerSlice(body{
				fieldTypeOf: valueOf.Field(fieldId).Type().Elem(),
				valueOf:     valueOf,
				typeOf:      valueOf.Type(),
				tables:      tables,
				fieldId:     fieldId,
				schema:      schema,
				mapp: &infosMap{
					pks:     pks,
					db:      db,
					tableId: tableId,
					addr:    addr,
				},
				driver: driver,
			}, helperAttribute)
			if err != nil {
				return err
			}
		case reflect.Struct:
			handlerStruct(body{
				fieldId:     fieldId,
				driver:      driver,
				fieldTypeOf: valueOf.Field(fieldId).Type(),
				valueOf:     valueOf,
				schema:      schema,
				mapp: &infosMap{
					pks:     pks,
					db:      db,
					tableId: tableId,
					addr:    addr,
				},
			}, newAttr)
		case reflect.Pointer:
			helperAttribute(body{
				fieldId:  fieldId,
				driver:   driver,
				nullable: true,
				tables:   tables,
				valueOf:  valueOf,
				typeOf:   valueOf.Type(),
				schema:   schema,
				mapp: &infosMap{
					pks:     pks,
					db:      db,
					tableId: tableId,
					addr:    addr,
				},
			})
		default:
			helperAttribute(body{
				fieldId: fieldId,
				driver:  driver,
				tables:  tables,
				valueOf: valueOf,
				typeOf:  valueOf.Type(),
				schema:  schema,
				mapp: &infosMap{
					pks:     pks,
					db:      db,
					tableId: tableId,
					addr:    addr,
				},
			})
		}
	}
	for i := range pks {
		addr := uintptr(valueOf.Field(fieldIds[i]).Addr().UnsafePointer())
		addrMap.set(addr, pks[i])
	}
	return nil
}

func handlerStruct(b body, create func(b body) error) error {
	return create(b)
}

func handlerSlice(b body, helper func(b body) error) error {
	switch b.fieldTypeOf.Kind() {
	case reflect.Uint8:
		return helper(b)
	}
	return nil
}

func newAttr(b body) error {
	at := createAtt(
		b.mapp.db,
		b.valueOf.Type().Field(b.fieldId).Name,
		b.schema,
		b.mapp.pks[0].tableName,
		b.mapp.tableId,
		b.fieldId,
		getTagValue(b.valueOf.Type().Field(b.fieldId).Tag.Get("goe"), "default:") != "",
		b.driver,
	)
	addrMap.set(b.mapp.addr, at)
	return nil
}

func getPks(typeOf reflect.Type) []reflect.StructField {
	var pks []reflect.StructField
	pks = append(pks, fieldsByTags("pk", typeOf)...)

	id, valid := getId(typeOf)
	if valid && !slices.ContainsFunc(pks, func(f reflect.StructField) bool { return f.Name == id.Name }) {
		pks = append(pks, id)
	}
	return pks
}

func getPk(db *DB, schema *string, valueOf reflect.Value, tableId int, driver model.Driver) ([]pk, []int, error) {
	typeOf := valueOf.Type()
	fields := getPks(typeOf)
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("goe: struct %q don't have a primary key setted", typeOf.Name())
	}

	table := utils.ParseTableNameByValue(valueOf)
	pks := make([]pk, len(fields))
	fieldIds := make([]int, len(fields))
	for i := range fields {
		fieldId := getFieldId(typeOf, fields[i].Name)
		pks[i] = createPk(db, schema, table, fields[i].Name, isReturningId(fields[i]), tableId, fieldId, driver)
		fieldIds[i] = fieldId
	}

	return pks, fieldIds, nil
}

func getFieldId(typeOf reflect.Type, fieldName string) int {
	for i := 0; i < typeOf.NumField(); i++ {
		if typeOf.Field(i).Name == fieldName {
			return i
		}
	}
	return 0
}

func isReturningId(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int") || getTagValue(id.Tag.Get("goe"), "default:") != ""
}

func isManyToOne(b body, createMany func(b body, typeOf reflect.Type) any, createOne func(b body, typeOf reflect.Type) any) any {
	fieldByName := b.tables.FieldByName(b.tableName).Elem()
	if fieldByName.IsValid() {
		for i := 0; i < fieldByName.NumField(); i++ {
			// check if there is a slice to typeOf
			if fieldByName.Field(i).Kind() == reflect.Slice {
				if fieldByName.Field(i).Type().Elem().Name() == b.typeOf.Name() {
					return createMany(b, fieldByName.Type())
				}
			}
		}
		if tableMtm := strings.ReplaceAll(b.typeOf.Name(), b.tableName, ""); tableMtm != b.typeOf.Name() {
			typeOfMtm := b.tables.FieldByName(tableMtm)
			if typeOfMtm.IsValid() && !typeOfMtm.IsZero() {
				typeOfMtm = typeOfMtm.Elem()
				for i := 0; i < typeOfMtm.NumField(); i++ {
					if typeOfMtm.Field(i).Kind() == reflect.Slice && typeOfMtm.Field(i).Type().Elem().Name() == b.tableName {
						return createMany(b, typeOfMtm.Field(i).Type().Elem())
					}
				}
			}
		}
		return createOne(b, fieldByName.Type())
	}
	return nil
}

func primaryKeys(str reflect.Type) (pks []reflect.StructField) {
	field, exists := getId(str)
	if exists {
		pks := make([]reflect.StructField, 1)
		pks[0] = field
		return pks
	} else {
		// TODO: Return anonymous pk para len(pks) == 0
		return fieldsByTags("pk", str)
	}
}

func fieldsByTags(tag string, str reflect.Type) (f []reflect.StructField) {
	f = make([]reflect.StructField, 0)

	for i := 0; i < str.NumField(); i++ {
		if strings.Contains(str.Field(i).Tag.Get("goe"), tag) {
			f = append(f, str.Field(i))
		}
	}
	return f
}

func getTagValue(FieldTag string, subTag string) string {
	values := strings.Split(FieldTag, ";")
	for _, v := range values {
		if after, found := strings.CutPrefix(v, subTag); found {
			return after
		}
	}
	return ""
}

func foreignKeyNamePattern(dbTables reflect.Value, fieldName string) (table, suffix string) {
	for r := 1; r <= len(fieldName); r++ {
		table = fieldName[:r]
		suffix = fieldName[r:]
		if dbTables.FieldByName(table).IsValid() {
			pks := getPks(dbTables.FieldByName(table).Elem().Type())
			for c := 1; c <= len(suffix); c++ {
				pkName := suffix[:c]
				if slices.ContainsFunc(pks, func(f reflect.StructField) bool {
					return f.Name == pkName
				}) {
					return table, pkName
				}
			}
		}
	}
	return "", ""
}

func helperAttribute(b body) error {
	table, prefix := foreignKeyNamePattern(b.tables, b.valueOf.Type().Field(b.fieldId).Name)
	if table != "" {
		b.stringInfos = stringInfos{prefixName: prefix, tableName: table, fieldName: b.valueOf.Type().Field(b.fieldId).Name}
		if mto := isManyToOne(b, createManyToOne, createOneToOne); mto != nil {
			switch v := mto.(type) {
			case manyToOne:
				if addrMap.get(b.mapp.addr) == nil {
					addrMap.set(b.mapp.addr, v)
				}
				for i := range b.mapp.pks {
					if !b.nullable && b.mapp.pks[i].fieldId == v.fieldId {
						b.mapp.pks[i].autoIncrement = false
					}
				}
			case oneToOne:
				if addrMap.get(b.mapp.addr) == nil {
					addrMap.set(b.mapp.addr, v)
				}
			}
			return nil
		}
	}
	newAttr(b)
	return nil
}

func getId(typeOf reflect.Type) (reflect.StructField, bool) {
	return typeOf.FieldByNameFunc(func(s string) bool {
		return strings.ToUpper(s) == "ID"
	})
}
