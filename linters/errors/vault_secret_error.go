package errors

import "fmt"

type VaultSecretError struct {
	Secret string
}

func NewVaultSecretError(secret string) VaultSecretError {
	return VaultSecretError{secret}
}

func (e VaultSecretError) Error() string {
	return fmt.Sprintf("'%s' is not a valid key", e.Secret)
}

func (e VaultSecretError) DocId() string {
	return "vault-secret-error"
}
