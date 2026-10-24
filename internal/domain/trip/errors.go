package trip

import "errors"

var (
	ErrDriverBusy    = errors.New("driver already has active trip")
	ErrNotFound      = errors.New("trip not found")
	ErrTripCompleted = errors.New("trip already ended")
)
