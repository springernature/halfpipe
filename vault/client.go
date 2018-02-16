package vault

type VaultClient interface {
	Exists(path string) (bool, error)
}