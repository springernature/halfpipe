package secret_resolver

import (
	"fmt"
)

type VaultSecretNotFoundError struct {
	prefix   string
	team     string
	pipeline string
	Secret   string
}

func NewVaultSecretNotFoundError(prefix string, team string, pipeline string, secret string) VaultSecretNotFoundError {
	return VaultSecretNotFoundError{
		prefix,
		team,
		pipeline,
		secret,
	}
}

func (e VaultSecretNotFoundError) Error() string {
	mapName, keyName := SecretToMapAndKey(e.Secret)

	path1 := fmt.Sprintf("/%s/%s/%s/%s", e.prefix, e.team, e.pipeline, mapName)
	path2 := fmt.Sprintf("/%s/%s/%s", e.prefix, e.team, mapName)

	return fmt.Sprintf("Could not find '%s' in '%s' or '%s'", keyName, path1, path2)
}

func (e VaultSecretNotFoundError) DocId() string {
	return "vault-secret-not-found"
}
