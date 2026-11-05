package decode

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ValidationErr
type ValidationErr struct {
	Err error
}

func NewVlidationErr(err error) error {
	return &ValidationErr{
		Err: err,
	}
}

func (e *ValidationErr) Error() string {
	return fmt.Sprintf("validation error: %v", e.Err)
}

func DecodeJSON(r *http.Request, data interface{}) error {
	if r.Body == nil {
		return NewVlidationErr(errors.New("request body is empmy"))
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		return NewVlidationErr(err)
	}
	return nil
}
