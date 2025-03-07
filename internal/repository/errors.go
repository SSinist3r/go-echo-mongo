package repository

import "errors"

var (
	// ErrNotFound is returned when a document is not found in the database
	ErrNotFound = errors.New("document not found")
)
