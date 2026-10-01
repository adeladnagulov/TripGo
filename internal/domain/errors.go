package domain

//думаю создать пакет errors
import "errors"

var ErrConflict = errors.New("entity already exists or violates unique constraint")
var ErrNotFound = errors.New("trip not found")
var ErrCannotFinishTrip = errors.New("trip already ended, or it not found")
