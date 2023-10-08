package usecase

import "errors"

// Used errors when working with data storage, for their processing in handlers.
var (
	ErrUniqueValue = errors.New("not unique value")
	ErrDeletedURL  = errors.New("URL mark on deleted")
	ErrNotFoundURL = errors.New("URL not found")
)
