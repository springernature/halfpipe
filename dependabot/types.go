package dependabot

var SupportedFiles = map[string]string{
	"Dockerfile":        "docker",
	"package-lock.json": "npm",
	"yarn.lock":         "npm",
	"Gemfile.lock":      "bundler",
	"pom.xml":           "maven",
	"build.gradle":      "gradle",
	"build.gradle.kt":   "gradle",
	"go.mod":            "gomod",
	"github-actions":    "github-actions",
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
