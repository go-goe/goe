package goe

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func init() {
	addrMap = &goeMap{mapField: make(map[uintptr]field)}
}

var ErrStructWithoutPrimaryKey = errors.New("goe")

// Open opens a database connection
//
// # Example
//
//	goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.Config{}))
func Open[T any](driver Driver) (*T, error) {
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
		if valueOf.Field(i).IsNil() {
			valueOf.Field(i).Set(reflect.ValueOf(reflect.New(valueOf.Field(i).Type().Elem()).Interface()))
		}
	}

	var err error
	// init Fields
	for tableId := range dbId {
		err = initField(valueOf, valueOf.Field(tableId).Elem(), dbTarget, tableId+1, driver)
		if err != nil {
			return nil, err
		}
	}

	err = driver.Init()
	if err != nil {
		return nil, err
	}

	dbTarget.Driver = driver
	return db, nil
}

func initField(tables reflect.Value, valueOf reflect.Value, db *DB, tableId int, driver Driver) error {
	pks, fieldIds, err := getPk(db, valueOf.Type(), tableId, driver)
	if err != nil {
		return err
	}

	for i := range pks {
		addrMap.set(uintptr(valueOf.Field(fieldIds[i]).Addr().UnsafePointer()), pks[i])
	}
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if slices.Contains(fieldIds, i) {
			table, prefix := checkTablePattern(tables, field)
			if table == "" && prefix == "" {
				continue
			}
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			err := handlerSlice(tables, valueOf.Field(i).Type().Elem(), valueOf, i, pks, db, tableId, i, driver)
			if err != nil {
				return err
			}
		case reflect.Struct:
			handlerStruct(valueOf.Field(i).Type(), valueOf, i, pks[0], db, tableId, i, driver)
		case reflect.Ptr:
			helperAttribute(tables, valueOf, i, db, tableId, i, driver, pks, true)
		default:
			helperAttribute(tables, valueOf, i, db, tableId, i, driver, pks, false)
		}
	}
	return nil
}

func handlerStruct(targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, tableId, fieldId int, driver Driver) {
	switch targetTypeOf.Name() {
	case "Time":
		newAttr(valueOf, i, p.tableName, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, fieldId, driver)
	}
}

func handlerSlice(tables reflect.Value, targetTypeOf reflect.Type, valueOf reflect.Value, i int, pks []*pk, db *DB, tableId, fieldId int, driver Driver) error {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		helperAttribute(tables, valueOf, i, db, tableId, fieldId, driver, pks, false)
	}
	return nil
}

func newAttr(valueOf reflect.Value, i int, table string, addr uintptr, db *DB, tableId, fieldId int, d Driver) {
	at := createAtt(
		db,
		valueOf.Type().Field(i).Name,
		table,
		tableId,
		fieldId,
		d,
	)
	addrMap.set(addr, at)
}

func getPk(db *DB, typeOf reflect.Type, tableId int, driver Driver) ([]*pk, []int, error) {
	var pks []*pk
	var fieldIds []int
	var fieldId int

	id, valid := typeOf.FieldByNameFunc(func(s string) bool {
		return strings.ToUpper(s) == "ID"
	})
	if valid {
		pks := make([]*pk, 1)
		fieldIds = make([]int, 1)
		fieldId = getFieldId(typeOf, id.Name)
		pks[0] = createPk(db, typeOf.Name(), id.Name, isAutoIncrement(id), tableId, fieldId, driver)
		fieldIds[0] = fieldId
		return pks, fieldIds, nil
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("%w: struct %q don't have a primary key setted", ErrStructWithoutPrimaryKey, typeOf.Name())
	}

	pks = make([]*pk, len(fields))
	fieldIds = make([]int, len(fields))
	for i := range fields {
		fieldId = getFieldId(typeOf, fields[i].Name)
		pks[i] = createPk(db, typeOf.Name(), fields[i].Name, isAutoIncrement(fields[i]), tableId, fieldId, driver)
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

func isAutoIncrement(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int")
}

func isManyToOne(db *DB, tables reflect.Value, typeOf reflect.Type, tableId, fieldId int, driver Driver, table, prefix, fieldName string) field {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createManyToOne(db, tables.Field(c).Elem().Type(), typeOf, tableId, fieldId, driver, prefix, fieldName)
					}
				}
			}
			if tableMtm := strings.ReplaceAll(typeOf.Name(), table, ""); tableMtm != typeOf.Name() {
				typeOfMtm := tables.FieldByName(tableMtm)
				if typeOfMtm.IsValid() && !typeOfMtm.IsZero() {
					typeOfMtm = typeOfMtm.Elem()
					for i := 0; i < typeOfMtm.NumField(); i++ {
						if typeOfMtm.Field(i).Kind() == reflect.Slice && typeOfMtm.Field(i).Type().Elem().Name() == table {
							return createManyToOne(db, typeOfMtm.Field(i).Type().Elem(), typeOf, tableId, fieldId, driver, prefix, fieldName)
						}
					}
				}
			}
			return createOneToOne(db, tables.Field(c).Elem().Type(), typeOf, tableId, fieldId, driver, prefix, fieldName)
		}
	}
	return nil
}

func primaryKeys(str reflect.Type) (pks []reflect.StructField) {
	Field, exists := str.FieldByName("Id")
	if exists {
		pks := make([]reflect.StructField, 1)
		pks[0] = Field
		return pks
	} else {
		//TODO: Return anonymous pk para len(pks) == 0
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

func checkTablePattern(tables reflect.Value, field reflect.StructField) (table, prefix string) {
	table, prefix = prefixNamePattern(tables, field)
	if table != "" {
		return table, prefix
	}
	return posfixNamePattern(tables, field)
}

func prefixNamePattern(tables reflect.Value, field reflect.StructField) (table, prefix string) {
	for r := len(field.Name) - 1; r > 1; r-- {
		if field.Name[r] < 'a' {
			table = field.Name[r:]
			prefix = field.Name[:r]
			if tables.FieldByName(table).IsValid() {
				return table, prefix
			}
		}
	}
	return "", ""
}

func posfixNamePattern(tables reflect.Value, field reflect.StructField) (table, prefix string) {
	for r := 0; r < len(field.Name); r++ {
		if field.Name[r] < 'a' {
			table = field.Name[:r]
			prefix = field.Name[r:]
			if tables.FieldByName(table).IsValid() {
				return table, prefix
			}
		}
	}
	return "", ""
}

func helperAttribute(tables reflect.Value, valueOf reflect.Value, i int, db *DB, tableId, fieldId int, driver Driver, pks []*pk, nullable bool) {
	table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
	if table != "" {
		if mto := isManyToOne(db, tables, valueOf.Type(), tableId, fieldId, driver, table, prefix, valueOf.Type().Field(i).Name); mto != nil {
			switch v := mto.(type) {
			case *manyToOne:
				ptr := uintptr(valueOf.Field(i).Addr().UnsafePointer())
				if v == nil {
					newAttr(valueOf, i, pks[0].tableName, ptr, db, tableId, fieldId, driver)
					return
				}
				if addrMap.get(ptr) == nil {
					addrMap.set(ptr, v)
					return
				}
				for _, pk := range pks {
					if !nullable && pk.fieldId == v.fieldId {
						pk.autoIncrement = false
					}
				}
			case *oneToOne:
				ptr := uintptr(valueOf.Field(i).Addr().UnsafePointer())
				if v == nil {
					newAttr(valueOf, i, pks[0].tableName, ptr, db, tableId, fieldId, driver)
					return
				}
				if addrMap.get(ptr) == nil {
					addrMap.set(ptr, v)
				}
			}
			return
		}
	}
	newAttr(valueOf, i, pks[0].tableName, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, fieldId, driver)
}
