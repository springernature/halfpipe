package linterrors

import "fmt"

type WrongHealthCheck struct {
	Path   string
	Reason string
}

func NewWrongHealthCheck(path string, reason string) WrongHealthCheck {
	return WrongHealthCheck{path, reason}
}

func (e WrongHealthCheck) Error() string {
	return fmt.Sprintf("invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
