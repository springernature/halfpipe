package secrets

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
)

var (
	errVaultAddrNotSet   = errors.NewVaultClientErrorf("Environment variable 'VAULT_ADDR' must be set!")
	errVaultTokenMissing = func(path string) error {
		return errors.NewVaultClientErrorf("Could not read vault token from path '%s'", path)
	}
	errVaultTokenInvalid = func(path string) error {
		return errors.NewVaultClientErrorf("Contents of '%s' does not look like a vault token!", path)
	}
)

type vaultClient struct {
	fs             afero.Afero
	getEnv         func(string) string
	vaultTokenPath string
}

func NewVaultClient(fs afero.Afero) vaultClient {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return vaultClient{
		fs:             fs,
		getEnv:         os.Getenv,
		vaultTokenPath: filepath.Join(homeDir, ".vault-token"),
	}
}

func (v vaultClient) Create() (vaultClient *api.Logical, err error) {
	if v.getEnv("VAULT_ADDR") == "" {
		err = errVaultAddrNotSet
		return
	}
	vaultTokenPath := v.vaultTokenPath
	token, err := v.fs.ReadFile(vaultTokenPath)
	if err != nil {
		err = errVaultTokenMissing(v.vaultTokenPath)
		return
	}
	if _, e := uuid.Parse(string(token)); e != nil {
		err = errVaultTokenInvalid(v.vaultTokenPath)
		return
	}
	apiClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return
	}

	apiClient.SetToken(string(token))
	vaultClient = apiClient.Logical()
	return
}
