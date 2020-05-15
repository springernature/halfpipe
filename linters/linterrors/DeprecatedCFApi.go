package linterrors

import "fmt"

type DeprecatedCFApiError struct {
	api string
}

func NewDeprecatedCFApiError(api string) DeprecatedCFApiError {
	return DeprecatedCFApiError{api}
}

func (e DeprecatedCFApiError) Error() string {
	return fmt.Sprintf("the Cloud Foundry instance at '%s' has been deprecated. Please see <http://status.ee.springernature.io/incidents/3ll7v596wznq>", e.api)
}
