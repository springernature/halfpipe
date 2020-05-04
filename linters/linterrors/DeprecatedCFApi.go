package linterrors

import "fmt"

type DeprecatedCFApiError struct {
	api string
}

func NewDeprecatedCFApiError(api string) DeprecatedCFApiError {
	return DeprecatedCFApiError{api}
}

func (e DeprecatedCFApiError) Error() string {
	return fmt.Sprintf("the Cloud Foundry instance at '%s' has been deprecated. Please see <https://ee-discourse.springernature.io/t/cloud-foundry-on-premises-deprecated/1292>", e.api)
}
