package domain

//думаю создать пакет errors
import "errors"

var ErrConflict = errors.New("entity already exists or violates unique constraint")
var ErrNotFound = errors.New("dsts not found")
