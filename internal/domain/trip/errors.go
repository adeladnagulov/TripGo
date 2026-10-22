package trip

import "errors"

var (
	ErrConflict         = errors.New("entity already exists or violates unique constraint")
	ErrNotFound         = errors.New("trip not found")
	ErrCannotFinishTrip = errors.New("trip already ended, or it not found")
)
