package errors

import "fmt"

type TooManyAppsError struct {
	Path   string
	Reason string
}

func NewTooManyAppsError(path string, reason string) TooManyAppsError {
	return TooManyAppsError{path, reason}
}

func (e TooManyAppsError) Error() string {
	return fmt.Sprintf("Invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
