package dependabot

var SupportedFiles = map[string]string{
	"Dockerfile":        "docker",
	"package-lock.json": "npm",
	"yarn.lock":         "npm",
	"Gemfile.lock":      "bundler",
}

type DependabotConfig struct {
	Depth         int
	Verbose       bool
	SkipFolders   []string
	SkipEcosystem []string
}

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
