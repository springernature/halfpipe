package errors

import (
	"fmt"
	"github.com/springernature/halfpipe/helpers"
)

type BadVaultSecretError struct {
	Secret string
}

func NewBadVaultSecretError(secret string) BadVaultSecretError {
	return BadVaultSecretError{secret}
}

func (e BadVaultSecretError) Error() string {
	return fmt.Sprintf("'%s' is not a valid key", e.Secret)
}

type NotFoundVaultSecretError struct {
	prefix   string
	team     string
	pipeline string
	Secret   string
}

func NewNotFoundVaultSecretError(prefix string, team string, pipeline string, secret string) NotFoundVaultSecretError {
	return NotFoundVaultSecretError{
		prefix,
		team,
		pipeline,
		secret,
	}
}

func (e NotFoundVaultSecretError) Error() string {
	mapName, keyName := helpers.SecretToMapAndKey(e.Secret)

	path1 := fmt.Sprintf("/%s/%s/%s/%s", e.prefix, e.team, e.pipeline, mapName)
	path2 := fmt.Sprintf("/%s/%s/%s", e.prefix, e.team, mapName)

	return fmt.Sprintf("Could not find '%s' in '%s' or '%s'", keyName, path1, path2)
}

type VaultClientError struct {
	message string
}

func NewVaultClientError(message string) VaultClientError {
	return VaultClientError{
		message: message,
	}
}

func (e VaultClientError) Error() string {
	return e.message
}
