package errors

import "fmt"

type NoRoutesError struct {
	Path   string
	Reason string
}

func NewNoRoutesError(path string, reason string) NoRoutesError {
	return NoRoutesError{path, reason}
}

func (e NoRoutesError) Error() string {
	return fmt.Sprintf("Invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
