package errors

import "fmt"

type VaultSecretError struct {
	Secret string
}

func NewVaultSecretError(secret string) VaultSecretError {
	return VaultSecretError{secret}
}

func (e VaultSecretError) Error() string {
	return fmt.Sprintf("'%s' is not a valid key, must be in format of ((mapName.keyName)) with allowed characters [a-zA-Z0-9-_]", e.Secret)
}
