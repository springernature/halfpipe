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
	Directory          string   `yaml:"directory"`
	Schedule           Schedule `yaml:"schedule"`
	Cooldown           Cooldown `yaml:"cooldown"`
	VersioningStrategy string   `yaml:"versioning-strategy"`
	Groups             Groups   `yaml:"groups"`
}

type MatchedPaths map[string]string

type Renderer func(paths MatchedPaths) Config
