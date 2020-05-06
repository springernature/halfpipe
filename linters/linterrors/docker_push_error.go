package linterrors

import "fmt"

type DockerPushError struct {
	Path   string
	Reason string
}

func NewDockerPushError(path string, reason string) DockerPushError {
	return DockerPushError{path, reason}
}

func (e DockerPushError) Error() string {
	return fmt.Sprintf("invalid CF Manifest: '%s': %s", e.Path, e.Reason)
}
