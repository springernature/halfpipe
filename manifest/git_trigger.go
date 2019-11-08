package manifest

import "strings"

type GitTrigger struct {
	Type          string
	URI           string   `yaml:"uri,omitempty"`
	BasePath      string   `yaml:"-"` //don't auto unmarshal
	PrivateKey    string   `yaml:"private_key,omitempty" secretAllowed:"true"`
	WatchedPaths  []string `yaml:"watched_paths,omitempty"`
	IgnoredPaths  []string `yaml:"ignored_paths,omitempty"`
	GitCryptKey   string   `yaml:"git_crypt_key,omitempty" secretAllowed:"true"`
	Branch        string   `yaml:"branch,omitempty"`
	Shallow       bool     `yaml:"shallow,omitempty"`
	ManualTrigger bool     `yaml:"manual_trigger,omitempty"`
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
