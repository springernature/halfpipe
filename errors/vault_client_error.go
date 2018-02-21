package errors

type VaultClientError struct {
	message string
}

func NewVaultClientError(message string) VaultClientError {
	return VaultClientError{
		message: message,
	}
}

func (e VaultClientError) Error() string {
	return e.message
}

func (e VaultClientError) DocId() string {
	return "vault-client-error"
}
