package dependabot

type Config struct {
	Version int
	Updates []Dependency
}

type Schedule struct {
	Interval string
}

type Cooldown struct {
	DefaultDays int `yaml:"default-days"`
}

type Group struct {
	UpdateTypes []string `yaml:"update-types"`
}

type Groups map[string]Group

// Dependency represents a single dependabot update entry.
// Field order matters for YAML output - fields are serialized in struct definition order.
type Dependency struct {
	PackageEcosystem   string   `yaml:"package-ecosystem"`
	Directories        []string `yaml:"directories"`
	Schedule           Schedule `yaml:"schedule"`
	Cooldown           Cooldown `yaml:"cooldown"`
	VersioningStrategy string   `yaml:"versioning-strategy,omitempty"`
	Groups             Groups   `yaml:"groups,omitempty"`
}

type MatchedPaths map[string]string

type Renderer func(paths MatchedPaths) Config
