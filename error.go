package goe

import "errors"

// ErrUniqueValue occurs on insert or update a uniquines value (e.g: primary key, unique columns...).
var ErrUniqueValue = errors.New("")

// ErrForeignKey occurs on insert or update a invalid foreign key.
var ErrForeignKey = errors.New("")

// ErrBadRequest is a error wrapper for any user interaction error (e.g: ErrUniqueValue, ErrForeignKey).
var ErrBadRequest = errors.New("")

// ErrNotFound occurs when the Find function returns zero results.
var ErrNotFound = errors.New("goe: not found any element on result set")
