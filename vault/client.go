package vault

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

func (v Vault) Exists(team string, pipeline string, mapKey string, keyName string) (bool, error) {
	return false, nil
}

func (v Vault) VaultPrefix() string {
	if v.prefix == "" {
		return "concourse"
	}
	return v.prefix
}
