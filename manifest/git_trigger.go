package manifest

import "strings"

// git trigger defines which git repo halfpipe will operate on. By convention
// there is always a git trigger as default. To disable it, set manual_trigger
// to true.
type GitTrigger struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Git repository URI. Defaults to the URI resolved from .git/config.
	URI      string `json:"uri,omitempty" yaml:"uri,omitempty"`
	BasePath string `json:"-" yaml:"-"` //don't auto unmarshal
	// SSH private key for cloning the repository.
	PrivateKey string `json:"private_key,omitempty" yaml:"private_key,omitempty" secretAllowed:"true"`
	// Only trigger when changes occur in these paths (globs supported). Paths should be relative to the repository root.
	WatchedPaths []string `json:"watched_paths,omitempty" yaml:"watched_paths,omitempty"`
	// Do not trigger when changes occur only in these paths (globs supported).
	IgnoredPaths []string `json:"ignored_paths,omitempty" yaml:"ignored_paths,omitempty"`
	// Base64-encoded git-crypt key to unlock an encrypted repository.
	GitCryptKey string `json:"git_crypt_key,omitempty" yaml:"git_crypt_key,omitempty" secretAllowed:"true"`
	// Branch to track. Required when running halfpipe on a non-default branch.
	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`
	// Shallow clone the repository (--depth 1). Defaults to false in Concourse and true in GitHub Actions.
	Shallow        bool `json:"shallow,omitempty" yaml:"shallow,omitempty"`
	ShallowDefined bool `json:"-" yaml:"-"` //don't auto unmarshal
	// Disable automatic triggering on commits.
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty" jsonschema:"default=false"`
}

func (git GitTrigger) GetTriggerAttempts() int {
	return 2
}

func (git GitTrigger) MarshalYAML() (any, error) {
	git.Type = "git"
	return git, nil
}

func (GitTrigger) GetTriggerName() string {
	return "git"
}

func (git GitTrigger) IsPublic() bool {
	return len(git.URI) > 4 && strings.HasPrefix(git.URI, "http")
}

func (git GitTrigger) GetRepoName() string {
	parts := strings.Split(git.URI, "/")
	repo := parts[len(parts)-1]

	return strings.Split(repo, ".git")[0]
}
