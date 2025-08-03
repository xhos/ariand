package db

import (
	"errors"
	"fmt"
)

// common database errors
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrConstraint   = errors.New("constraint violation")
	ErrTransaction  = errors.New("transaction failed")
)

// error wrapping helpers
func NotFoundError(resource string, id interface{}) error {
	return fmt.Errorf("%s with id %v: %w", resource, id, ErrNotFound)
}

func ConflictError(resource string, field string, value interface{}) error {
	return fmt.Errorf("%s %s=%v already exists: %w", resource, field, value, ErrConflict)
}

func ConstraintError(constraint string) error {
	return fmt.Errorf("constraint %s violated: %w", constraint, ErrConstraint)
}
