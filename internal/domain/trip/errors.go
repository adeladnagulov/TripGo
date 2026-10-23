package trip

import "errors"

var (
	ErrDriverBusy       = errors.New("driver already has active trip")
	ErrNotFound         = errors.New("trip not found")
	ErrCannotFinishTrip = errors.New("trip already ended, or it not found")
)
