package aggregate

import "github.com/olauro/goe/query"

func Count(t any) *query.Count {
	return &query.Count{Field: t}
}
