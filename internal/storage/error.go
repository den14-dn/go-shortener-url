package storage

import "errors"

// Description of the errors used when working with the data warehouse.
var (
	ErrUniqueValue = errors.New("not unique value")
	ErrDeletedURL  = errors.New("URL mark on deleted")
	ErrNotFoundURL = errors.New("URL not found")
)
