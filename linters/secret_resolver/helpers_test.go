package secret_resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretsToMapSuccessfull(t *testing.T) {
	secretStr := "((foo.key))"
	name, key := SecretToMapAndKey(secretStr)

	assert.Equal(t, "foo", name)
	assert.Equal(t, "key", key)
}
