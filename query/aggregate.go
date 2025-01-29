package query

import (
	"fmt"
)

type Count struct {
	Field any
	Value int64
}

func (c *Count) Scan(src any) error {
	v, ok := src.(int64)
	if !ok {
		return fmt.Errorf("error scan aggregate")
	}

	c.Value = v
	return nil
}

func (c *Count) Aggregate(s string) string {
	return fmt.Sprintf("COUNT(%v)", s)
}

func GetCount(t any) *Count {
	return &Count{Field: t}
}
