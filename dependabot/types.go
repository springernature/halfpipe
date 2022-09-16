package dependabot

type Config struct {
	Version int
	Updates []Dependency
}

type Schedule struct {
	Interval string
}

type Dependency struct {
	PackageEcosystem string `yaml:"package-ecosystem"`
	Directory        string
	Schedule         Schedule
}

type MatchedPaths map[string]string
