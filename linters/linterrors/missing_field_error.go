package linterrors

import "fmt"

type MissingFieldError struct {
	Name string
}

func NewMissingField(name string) MissingFieldError {
	return MissingFieldError{name}
}

func (e MissingFieldError) Error() string {
	return fmt.Sprintf("missing field: '%s'", e.Name)
}
