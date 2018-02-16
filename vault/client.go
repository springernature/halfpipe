package vault

type VaultClient interface {
	Exists(prefix string, team string, pipeline string, mapKey string, keyName string) (bool, error)
}

type Vault struct{}

func (Vault) Exists(prefix string, team string, pipeline string, mapKey string, keyName string) (bool, error) {
	return false, nil
}
