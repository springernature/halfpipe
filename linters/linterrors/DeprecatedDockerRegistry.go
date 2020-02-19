package linterrors

import "fmt"

type DeprecatedDockerRegistryError struct {
	hostname string
}

func NewDeprecatedDockerRegistryError(hostname string) DeprecatedDockerRegistryError {
	return DeprecatedDockerRegistryError{hostname}
}

func (e DeprecatedDockerRegistryError) Error() string {
	return fmt.Sprintf("the docker registry '%s' has been deprecated. More information: https://ee-discourse.springernature.io/c/news-updates", e.hostname)
}
