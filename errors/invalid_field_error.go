package errors

import "fmt"

type InvalidFieldError struct {
	Name   string
	Reason string
}

func NewInvalidField(name string, reason string) InvalidFieldError {
	return InvalidFieldError{name, reason}
}

func (e InvalidFieldError) Error() string {
	return fmt.Sprintf("Invalid value for '%s': %s", e.Name, e.Reason)
}
