package goe

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

var ErrStructWithoutPrimaryKey = errors.New("goe")

func Open[T any](driver Driver, config Config) (*T, error) {
	db := new(T)
	valueOf := reflect.ValueOf(db)
	if valueOf.Kind() != reflect.Ptr {
		return nil, errors.New("goe: the target value needs to be pass as a pointer")
	}
	dbTarget := new(DB)
	valueOf = valueOf.Elem()

	if addrMap == nil {
		addrMap = make(map[uintptr]field)
	}

	// set value for Fields
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).IsNil() {
			valueOf.Field(i).Set(reflect.ValueOf(reflect.New(valueOf.Field(i).Type().Elem()).Interface()))
		}
	}

	var err error
	var tableId uint = 1
	// init Fields
	for i := 0; i < valueOf.NumField(); i++ {
		err = initField(valueOf, valueOf.Field(i).Elem(), dbTarget, tableId, driver)
		if err != nil {
			return nil, err
		}
		tableId++
	}

	dbTarget.Driver = driver
	dbTarget.Driver.Init(dbTarget)
	dbTarget.Config = &config
	return db, nil
}

func initField(tables reflect.Value, valueOf reflect.Value, db *DB, tableId uint, driver Driver) error {
	pks, FieldNames, err := getPk(db, valueOf.Type(), tableId, driver)
	if err != nil {
		return err
	}

	for i := range pks {
		addrMap[uintptr(valueOf.FieldByName(FieldNames[i]).Addr().UnsafePointer())] = pks[i]
	}
	var Field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		Field = valueOf.Type().Field(i)
		//skip primary key
		if slices.Contains(FieldNames, Field.Name) {
			//TODO: Check this
			table, prefix := checkTablePattern(tables, Field)
			if table == "" && prefix == "" {
				continue
			}
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			err := handlerSlice(tables, valueOf.Field(i).Type().Elem(), valueOf, i, pks, db, tableId, driver)
			if err != nil {
				return err
			}
		case reflect.Struct:
			handlerStruct(valueOf.Field(i).Type(), valueOf, i, pks[0], db, tableId, driver)
		case reflect.Ptr:
			helperAttribute(tables, valueOf, i, db, tableId, driver, pks, true)
		default:
			helperAttribute(tables, valueOf, i, db, tableId, driver, pks, false)
		}
	}
	return nil
}

func handlerStruct(targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, tableId uint, driver Driver) {
	switch targetTypeOf.Name() {
	case "Time":
		newAttr(valueOf, i, p.tableBytes, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, driver)
	}
}

func handlerSlice(tables reflect.Value, targetTypeOf reflect.Type, valueOf reflect.Value, i int, pks []*pk, db *DB, tableId uint, driver Driver) error {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		helperAttribute(tables, valueOf, i, db, tableId, driver, pks, false)
	}
	return nil
}

func newAttr(valueOf reflect.Value, i int, tableBytes []byte, addr uintptr, db *DB, tableId uint, d Driver) {
	at := createAtt(
		db,
		valueOf.Type().Field(i).Name,
		tableBytes,
		tableId,
		d,
	)
	addrMap[addr] = at
}

func getPk(db *DB, typeOf reflect.Type, tableId uint, driver Driver) ([]*pk, []string, error) {
	var pks []*pk
	var FieldsNames []string

	id, valid := typeOf.FieldByName("Id")
	if valid {
		pks := make([]*pk, 1)
		FieldsNames = make([]string, 1)
		pks[0] = createPk(db, []byte(typeOf.Name()), id.Name, isAutoIncrement(id), tableId, driver)
		FieldsNames[0] = id.Name
		return pks, FieldsNames, nil
	}

	Fields := fieldsByTags("pk", typeOf)
	if len(Fields) == 0 {
		return nil, nil, fmt.Errorf("%w: struct %q don't have a primary key setted", ErrStructWithoutPrimaryKey, typeOf.Name())
	}

	pks = make([]*pk, len(Fields))
	FieldsNames = make([]string, len(Fields))
	for i := range Fields {
		pks[i] = createPk(db, []byte(typeOf.Name()), Fields[i].Name, isAutoIncrement(Fields[i]), tableId, driver)
		FieldsNames[i] = Fields[i].Name
	}

	return pks, FieldsNames, nil
}

func isAutoIncrement(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int")
}

func isManyToOne(db *DB, tables reflect.Value, typeOf reflect.Type, tableId uint, driver Driver, table, prefix string) field {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createManyToOne(db, tables.Field(c).Elem().Type(), typeOf, tableId, driver, prefix)
					}
				}
			}
			if tableMtm := strings.ReplaceAll(typeOf.Name(), table, ""); tableMtm != typeOf.Name() {
				typeOfMtm := tables.FieldByName(tableMtm)
				if typeOfMtm.IsValid() && !typeOfMtm.IsZero() {
					typeOfMtm = typeOfMtm.Elem()
					for i := 0; i < typeOfMtm.NumField(); i++ {
						if typeOfMtm.Field(i).Kind() == reflect.Slice && typeOfMtm.Field(i).Type().Elem().Name() == table {
							return createManyToOne(db, typeOfMtm.Field(i).Type().Elem(), typeOf, tableId, driver, prefix)
						}
					}
				}
			}
			return createOneToOne(db, tables.Field(c).Elem().Type(), typeOf, tableId, driver, prefix)
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

func checkTablePattern(tables reflect.Value, Field reflect.StructField) (table, prefix string) {
	table = getTagValue(Field.Tag.Get("goe"), "table:")
	if table != "" {
		prefix = strings.ReplaceAll(Field.Name, table, "")
		return table, prefix
	}
	if table == "" {
		for r := len(Field.Name) - 1; r > 1; r-- {
			if Field.Name[r] < 'a' {
				table = Field.Name[r:]
				prefix = Field.Name[:r]
				if tables.FieldByName(table).IsValid() {
					return table, prefix
				}
			}
		}
		if !tables.FieldByName(table).IsValid() {
			table = ""
		}
	}
	return table, prefix
}

func helperAttribute(tables reflect.Value, valueOf reflect.Value, i int, db *DB, tableId uint, driver Driver, pks []*pk, nullable bool) {
	table, prefix := checkTablePattern(tables, valueOf.Type().Field(i))
	if table != "" {
		if mto := isManyToOne(db, tables, valueOf.Type(), tableId, driver, table, prefix); mto != nil {
			switch v := mto.(type) {
			case *manyToOne:
				if v == nil {
					newAttr(valueOf, i, pks[0].tableBytes, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, driver)
					break
				}
				addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
				for _, pk := range pks {
					if !nullable && pk.structAttributeName == v.structAttributeName {
						pk.autoIncrement = false
						v.primaryKey = true
					}
				}
			case *oneToOne:
				if v == nil {
					newAttr(valueOf, i, pks[0].tableBytes, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, driver)
					break
				}
				addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
				for _, pk := range pks {
					//TODO: Check this
					if !nullable && pk.structAttributeName == v.structAttributeName {
						pk.autoIncrement = false
						v.primaryKey = true
					}
				}
			}
			return
		}
	}
	newAttr(valueOf, i, pks[0].tableBytes, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, tableId, driver)
}
