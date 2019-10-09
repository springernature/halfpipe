package linterrors

import "fmt"

type NoNameError struct {
	Path   string
	Reason string
}

func NewNoNameError(path string, reason string) NoNameError {
	return NoNameError{path, reason}
}

func (e NoNameError) Error() string {
	return fmt.Sprintf("Invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
