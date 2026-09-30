package domain

import "errors"

var ErrConflict = errors.New("entity already exists or violates unique constraint")
