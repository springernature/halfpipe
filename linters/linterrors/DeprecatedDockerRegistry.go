package linterrors

import "fmt"

type DeprecatedDockerRegistryError struct {
	hostname string
}

func NewDeprecatedDockerRegistryError(hostname string) DeprecatedDockerRegistryError {
	return DeprecatedDockerRegistryError{hostname}
}

func (e DeprecatedDockerRegistryError) Error() string {
	return fmt.Sprintf("the docker registry '%s' has been deprecated. Please see <http://status.ee.springernature.io/incidents/bl8y88pmcz23>", e.hostname)
}
