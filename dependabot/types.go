package dependabot

type Registry struct {
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

type Config struct {
	Version    int                 `yaml:"version"`
	Registries map[string]Registry `yaml:"registries,omitempty"`
	Updates    []Dependency        `yaml:"updates"`
}

type Schedule struct {
	Interval string
}

type Cooldown struct {
	DefaultDays int `yaml:"default-days"`
}

type Group struct {
	Patterns    []string `yaml:"patterns,omitempty"`
	UpdateTypes []string `yaml:"update-types,omitempty"`
}

type Groups map[string]Group

type CommitMessage struct {
	Prefix  string `yaml:"prefix"`
	Include string `yaml:"include"`
}

type Ignore struct {
	DependencyName string   `yaml:"dependency-name"`
	UpdateTypes    []string `yaml:"update-types,omitempty"`
}

// Dependency represents a single dependabot update entry.
// Field order matters for YAML output - fields are serialized in struct definition order.
type Dependency struct {
	PackageEcosystem      string        `yaml:"package-ecosystem"`
	Directories           []string      `yaml:"directories"`
	Schedule              Schedule      `yaml:"schedule"`
	Cooldown              Cooldown      `yaml:"cooldown"`
	OpenPullRequestsLimit int           `yaml:"open-pull-requests-limit"`
	Labels                []string      `yaml:"labels"`
	CommitMessage         CommitMessage `yaml:"commit-message"`
	VersioningStrategy    string        `yaml:"versioning-strategy,omitempty"`
	Groups                Groups        `yaml:"groups,omitempty"`
	Ignore                []Ignore      `yaml:"ignore,omitempty"`
	Registries            []string      `yaml:"registries,omitempty"`
}

type MatchedPaths map[string]string

type Renderer func(paths MatchedPaths) Config
