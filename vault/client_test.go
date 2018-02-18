package vault

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorWhenVaultAddrNotSet(t *testing.T) {
	os.Unsetenv(api.EnvVaultAddress)
	client := NewVaultClient("")
	_, err := client.Exists("", "", "", "")

	assert.IsType(t, errors.VaultClientError{}, err)
}
