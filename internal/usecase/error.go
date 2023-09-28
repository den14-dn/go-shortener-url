package usecase

import "errors"

var (
	ErrUniqueValue = errors.New("not unique value")
	ErrDeletedURL  = errors.New("URL mark on deleted")
	ErrNotFoundURL = errors.New("URL not found")
)
