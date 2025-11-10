package goe

import "errors"

var ErrUniqueValue = errors.New("")
var ErrForeignKey = errors.New("")
var ErrBadRequest = errors.New("")
var ErrNotFound = errors.New("goe: not found any element on result set")
