package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/atc/creds/vault"
	"github.com/hashicorp/vault/api"
	"github.com/springernature/halfpipe/errors"
)

type Client interface {
	Exists(team string, pipeline string, mapKey string, keyName string) (bool, error)
	VaultPrefix() string
}

type Vault struct {
	prefix string
}

func NewVaultClient(prefix string) Vault {
	return Vault{prefix}
}

func (v Vault) Exists(team string, pipeline string, mapKey string, keyName string) (foundValue bool, error error) {
	client, err := v.createRestClient()
	if err != nil {
		error = err
		return
	}

	vault := vault.Vault{
		VaultClient:  client,
		PathPrefix:   v.prefix,
		TeamName:     team,
		PipelineName: pipeline,
	}

	data, found, err := vault.Get(template.VariableDefinition{Name: mapKey})
	if err != nil || !found {
		foundValue = found
		error = err
		return
	}

	_, foundValue = data.(map[interface{}]interface{})[keyName]
	return
}

func (v Vault) VaultPrefix() string {
	if v.prefix == "" {
		return "concourse"
	}
	return v.prefix
}

func (v Vault) createRestClient() (client *api.Logical, error error) {
	if os.Getenv(api.EnvVaultAddress) == "" {
		error = errors.NewVaultClientError("Required env var 'VAULT_ADDR' not set")
		return
	}

	config := api.DefaultConfig()
	c, err := api.NewClient(config)
	if err != nil {
		error = err
		return
	}

	token, err := v.readToken()
	if err != nil {
		error = err
		return
	}

	c.SetToken(token)
	client = c.Logical()
	return
}

func (v Vault) readToken() (token string, error error) {
	user, err := user.Current()
	if err != nil {
		error = err
		return
	}

	b, error := ioutil.ReadFile(fmt.Sprintf("%s/.vault-token", user.HomeDir))
	token = string(b)
	return
}
