package linterrors

import "fmt"

type BadRoutesError struct {
	Path   string
	Reason string
}

func NewBadRoutesError(path string, reason string) BadRoutesError {
	return BadRoutesError{path, reason}
}

func (e BadRoutesError) Error() string {
	return fmt.Sprintf("invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
