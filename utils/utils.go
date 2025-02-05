package utils

import "strings"

func TableNamePattern(name string) string {
	name += "s"
	return strings.ToLower(name)
}

func ColumnNamePattern(name string) string {
	return strings.ToLower(name)
}
