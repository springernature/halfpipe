package manifest

import "strings"

type GitTrigger struct {
	Type           string
	URI            string   `json:"uri,omitempty" yaml:"uri,omitempty"`
	BasePath       string   `json:"-" yaml:"-"` //don't auto unmarshal
	PrivateKey     string   `json:"private_key,omitempty" yaml:"private_key,omitempty" secretAllowed:"true"`
	WatchedPaths   []string `json:"watched_paths,omitempty" yaml:"watched_paths,omitempty"`
	IgnoredPaths   []string `json:"ignored_paths,omitempty" yaml:"ignored_paths,omitempty"`
	GitCryptKey    string   `json:"git_crypt_key,omitempty" yaml:"git_crypt_key,omitempty" secretAllowed:"true"`
	Branch         string   `json:"branch,omitempty" yaml:"branch,omitempty"`
	Shallow        bool     `json:"shallow,omitempty" yaml:"shallow,omitempty"`
	ShallowDefined bool     `json:"-" yaml:"-"` //don't auto unmarshal
	ManualTrigger  bool     `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
}

func (git GitTrigger) GetTriggerAttempts() int {
	return 2
}

func (git GitTrigger) MarshalYAML() (interface{}, error) {
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
