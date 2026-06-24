package utils

import (
	"reflect"
	"strings"
	"unicode"
)

func ParseTableNameByValue(valueOf reflect.Value) string {
	if tableName := TableNameMethod(valueOf); tableName != "" {
		return tableName
	}
	return TableNamePattern(valueOf.Type().Name())
}

func ParseTableNameByType(typeOf reflect.Type) string {
	valueOf := reflect.New(typeOf)
	if tableName := TableNameMethod(valueOf); tableName != "" {
		return tableName
	}
	return TableNamePattern(typeOf.Name())
}

func TableNameMethod(valueOf reflect.Value) string {
	var method reflect.Value
	if method = valueOf.MethodByName("TableName"); method.IsValid() {
		if method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
			return method.Call(nil)[0].String()
		}
	}
	if valueOf.Type().Kind() == reflect.Struct && valueOf.Addr().IsValid() {
		if method = valueOf.Addr().MethodByName("TableName"); method.IsValid() {
			if method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
				return method.Call(nil)[0].String()
			}
		}
	}
	return ""
}

// TableNamePattern is the default name patterning for mapping struct to table
func TableNamePattern(name string) string {
	if len(name) == 0 {
		return name
	}
	name = ToSnakeCase(name)
	if name[len(name)-1] != 's' {
		name += "s"
	}
	return name
}

// ColumnNamePattern is the default name patterning for mapping struct fields to table columns
func ColumnNamePattern(name string) string {
	if len(name) == 0 {
		return name
	}
	return ToSnakeCase(name)
}

func ToSnakeCase(name string) string {
	if len(name) == 0 {
		return name
	}

	result := strings.Builder{}
	for i := 0; i < len(name); i++ {
		letter := rune(name[i])
		if unicode.IsUpper(letter) {
			letter = unicode.ToLower(letter)
			if i > 0 {
				prevLetter := rune(name[i-1])
				if unicode.IsLower(prevLetter) {
					result.WriteRune('_')
				} else if i+1 < len(name) && unicode.IsLower(rune(name[i+1])) {
					result.WriteRune('_')
				}
			}
		}
		result.WriteRune(letter)
	}

	return result.String()
}

// IsFieldHasSchema check if field has schema tag or schema suffix
func IsFieldHasSchema(valueOf reflect.Value, i int) bool {
	return strings.Contains(valueOf.Type().Field(i).Tag.Get("goe"), "schema") ||
		strings.HasSuffix(valueOf.Field(i).Elem().Type().Name(), "Schema")
}
