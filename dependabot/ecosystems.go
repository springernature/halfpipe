package dependabot

// ecosystemConfig holds the complete configuration for a supported ecosystem,
// including which files indicate its presence and how to render its dependabot entry.
type ecosystemConfig struct {
	files              []string // filenames that indicate this ecosystem
	versioningStrategy string   // empty means omit from output
	groups             Groups   // nil means omit from output
	registries         []string // registry names to reference; nil means omit from output
}

// registryDefinitions defines the private registries that dependabot should use.
var registryDefinitions = map[string]Registry{
	"sn-artifactory": {
		Type:     "maven-repository",
		URL:      "https://springernature.jfrog.io/artifactory/libs-release/",
		Username: "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
		Password: "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	},
}

// semverGroups is the standard groups config for ecosystems that support SemVer.
var semverGroups = Groups{
	"minor-and-patch": Group{
		UpdateTypes: []string{"minor", "patch"},
	},
}

// ecosystems defines all supported ecosystems, their indicator files,
// and their dependabot rendering defaults.
var ecosystems = map[string]ecosystemConfig{
	"bun":            {files: []string{"bun.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"bundler":        {files: []string{"Gemfile.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"cargo":          {files: []string{"Cargo.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"composer":       {files: []string{"composer.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"docker":         {files: []string{"Dockerfile"}},
	"docker-compose": {files: []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}},
	"elm":            {files: []string{"elm.json"}, versioningStrategy: "increase", groups: semverGroups},
	"github-actions": {}, // detected via .github/workflows prefix, not by filename
	"gomod":          {files: []string{"go.mod"}, groups: semverGroups},
	"gradle":         {files: []string{"build.gradle", "build.gradle.kt"}, groups: semverGroups, registries: []string{"sn-artifactory"}},
	"helm":           {files: []string{"Chart.yaml"}},
	"maven":          {files: []string{"pom.xml"}, groups: semverGroups, registries: []string{"sn-artifactory"}},
	"mix":            {files: []string{"mix.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"npm":            {files: []string{"package-lock.json", "yarn.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"nuget":          {files: []string{"packages.config"}, groups: semverGroups},
	"pip":            {files: []string{"requirements.txt", "Pipfile.lock", "setup.py", "setup.cfg", "pyproject.toml"}, versioningStrategy: "increase", groups: semverGroups},
	"pub":            {files: []string{"pubspec.lock"}, versioningStrategy: "increase", groups: semverGroups},
	"swift":          {files: []string{"Package.resolved"}, groups: semverGroups},
	"terraform":      {files: []string{"main.tf", "versions.tf", "providers.tf", "variables.tf", "outputs.tf", "terraform.tf"}},
	"uv":             {files: []string{"uv.lock"}, versioningStrategy: "increase", groups: semverGroups},
}

// supportedFiles builds a filename-to-ecosystem lookup map from the ecosystems config.
func supportedFiles() map[string]string {
	m := map[string]string{}
	for ecosystem, cfg := range ecosystems {
		for _, file := range cfg.files {
			m[file] = ecosystem
		}
	}
	return m
}
