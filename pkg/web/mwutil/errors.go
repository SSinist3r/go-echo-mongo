package mwutil

import "errors"

var (
	// ErrInvalidAPIKey is returned when the provided API key is invalid
	ErrInvalidAPIKey = errors.New("invalid api key")

	// ErrMissingAPIKey is returned when no API key is provided
	ErrMissingAPIKey = errors.New("missing api key")
)
