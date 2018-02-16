package vault

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe/errors"
	"os"
	"github.com/hashicorp/vault/api"
)

func TestErrorWhenVaultAddrNotSet(t *testing.T) {
	os.Unsetenv(api.EnvVaultAddress)
	client := NewVaultClient("")
	_, err := client.Exists("", "", "", "")

	assert.IsType(t, errors.VaultClientError{}, err)
}
