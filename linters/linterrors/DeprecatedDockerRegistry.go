package linterrors

import "fmt"

type DeprecatedDockerRegistryError struct {
	hostname string
}

func NewDeprecatedDockerRegistryError(hostname string) DeprecatedDockerRegistryError {
	return DeprecatedDockerRegistryError{hostname}
}

func (e DeprecatedDockerRegistryError) Error() string {
	return fmt.Sprintf("The docker registry '%s' has been deprecated. Please see <https://ee-discourse.springernature.io/t/internal-docker-registries-end-of-life/1317>", e.hostname)
}
