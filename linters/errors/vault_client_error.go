package errors

import "fmt"

type VaultClientError struct {
	message string
}

func NewVaultClientErrorf(format string, a ...interface{}) VaultClientError {
	return VaultClientError{
		message: fmt.Sprintf(format, a...),
	}
}

func (e VaultClientError) Error() string {
	return e.message
}
