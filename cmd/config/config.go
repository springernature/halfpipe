package config

var (
	// These field will be populated in Concourse
	// go build -ldflags "-X main.version=..."
	Version    string
	CompiledAt string
	GitCommit  string

	DocHost     string
	VaultPrefix string
)
