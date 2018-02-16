package errors

import (
	"fmt"
	"github.com/springernature/halfpipe/helpers"
)

type BadSecret struct {
	Secret string
}

func NewBadVaultSecretError(key string) BadSecret {
	return BadSecret{key}
}

func (e BadSecret) Error() string {
	return fmt.Sprintf("'%s' is not a valid key", e.Secret)
}

type NotFoundVaultSecretError struct {
	Secret string
}

func NewNotFoundVaultSecretError(secret string) NotFoundVaultSecretError {
	return NotFoundVaultSecretError{
		Secret: secret,
	}
}

func (e NotFoundVaultSecretError) Error() string {
	mapName, keyName := helpers.SecretToMapAndKey(e.Secret)

	path1 := fmt.Sprintf("/springernature/team/pipeline/%s", mapName)
	path2 := fmt.Sprintf("/springernature/team/%s", mapName)

	return fmt.Sprintf("Could not find '%s' under '%s' or '%s'", keyName, path1, path2)
}
