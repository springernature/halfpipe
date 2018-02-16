package vault

type VaultClient interface {
	Exists(path string) (bool, error)
}

type Vault struct{}

func (Vault) Exists(path string) (bool, error) {
	return false, nil
}
