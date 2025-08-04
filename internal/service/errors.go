package service

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrPermission    = errors.New("permission denied")
	ErrConflict      = errors.New("conflict")
	ErrValidation    = errors.New("validation failed")
	ErrUnimplemented = errors.New("unimplemented")
)

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}

	// if it's already a sentinel, preserve it
	var sentinel error
	switch {
	case errors.Is(err, ErrNotFound):
		sentinel = ErrNotFound
	case errors.Is(err, ErrPermission):
		sentinel = ErrPermission
	case errors.Is(err, ErrConflict):
		sentinel = ErrConflict
	case errors.Is(err, ErrValidation):
		sentinel = ErrValidation
	case errors.Is(err, ErrUnimplemented):
		sentinel = ErrUnimplemented
	}

	if sentinel != nil {
		return fmt.Errorf("%s: %w", op, sentinel)
	}

	return fmt.Errorf("%s: %w", op, err)
}
