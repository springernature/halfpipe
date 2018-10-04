package manifest

import "strings"

type TaskList []Task

type Task interface {
	ReadsFromArtifacts() bool
	GetAttempts() int
	SavesArtifacts() bool
}

type Manifest struct {
	Team            string
	Pipeline        string
	SlackChannel    string         `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TriggerInterval string         `json:"trigger_interval" yaml:"trigger_interval,omitempty"`
	Repo            Repo           `yaml:"repo,omitempty"`
	FeatureToggles  FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	Tasks           TaskList
}

type Repo struct {
	URI          string   `json:"uri,omitempty" yaml:"uri,omitempty"`
	BasePath     string   `json:"-" yaml:"-"` //don't auto unmarshal
	PrivateKey   string   `json:"private_key,omitempty" yaml:"private_key,omitempty"`
	WatchedPaths []string `json:"watched_paths,omitempty" yaml:"watched_paths,omitempty"`
	IgnoredPaths []string `json:"ignored_paths,omitempty" yaml:"ignored_paths,omitempty"`
	GitCryptKey  string   `json:"git_crypt_key,omitempty" yaml:"git_crypt_key,omitempty"`
	Branch       string   `json:"branch,omitempty" yaml:"branch,omitempty"`
}

func (repo Repo) IsPublic() bool {
	return len(repo.URI) > 4 && repo.URI[:4] == "http"
}

type Docker struct {
	Image    string
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Run struct {
	Type             string
	Name             string
	ManualTrigger    bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Script           string
	Docker           Docker
	Vars             Vars     `yaml:"vars,omitempty"`
	SaveArtifacts    []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	RestoreArtifacts bool     `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Parallel         bool     `yaml:"parallel,omitempty"`
	Retries          int      `yaml:"retries,omitempty"`
}

func (r Run) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r Run) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r Run) GetAttempts() int {
	return 1 + r.Retries
}

type DockerPush struct {
	Type             string
	Name             string
	ManualTrigger    bool `json:"manual_trigger" yaml:"manual_trigger"`
	Username         string
	Password         string
	Image            string
	Vars             Vars
	RestoreArtifacts bool `json:"restore_artifacts" yaml:"restore_artifacts"`
	Parallel         bool `yaml:"parallel,omitempty"`
	Retries          int  `yaml:"retries,omitempty"`
}

func (r DockerPush) SavesArtifacts() bool {
	return false
}

func (r DockerPush) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r DockerPush) GetAttempts() int {
	return 1 + r.Retries
}

type DockerCompose struct {
	Type             string
	Name             string
	Command          string
	ManualTrigger    bool `json:"manual_trigger" yaml:"manual_trigger"`
	Vars             Vars
	Service          string
	SaveArtifacts    []string `json:"save_artifacts"`
	RestoreArtifacts bool     `json:"restore_artifacts" yaml:"restore_artifacts"`
	Parallel         bool     `yaml:"parallel,omitempty"`
	Retries          int      `yaml:"retries,omitempty"`
}

func (r DockerCompose) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r DockerCompose) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r DockerCompose) GetAttempts() int {
	return 1 + r.Retries
}

type DeployCF struct {
	Type           string
	Name           string
	ManualTrigger  bool `json:"manual_trigger" yaml:"manual_trigger"`
	API            string
	Space          string
	Org            string
	Username       string
	Password       string
	Manifest       string
	TestDomain     string `json:"test_domain" yaml:"test_domain"`
	Vars           Vars
	DeployArtifact string   `json:"deploy_artifact"`
	PrePromote     TaskList `json:"pre_promote"`
	Parallel       bool     `yaml:"parallel,omitempty"`
	Timeout        string
	Retries        int `yaml:"retries,omitempty"`
}

func (r DeployCF) SavesArtifacts() bool {
	return false
}

func (r DeployCF) ReadsFromArtifacts() bool {
	return r.DeployArtifact != "" || strings.HasPrefix(r.Manifest, "../artifacts/")
}

func (r DeployCF) GetAttempts() int {
	return 2 + r.Retries
}

type ConsumerIntegrationTest struct {
	Type                 string
	Name                 string
	Consumer             string
	ConsumerHost         string `json:"consumer_host" yaml:"consumer_host"`
	ProviderHost         string `json:"provider_host" yaml:"provider_host"`
	Script               string
	DockerComposeService string `json:"docker_compose_service" yaml:"docker_compose_service"`
	Parallel             bool   `yaml:"parallel,omitempty"`
	Vars                 Vars
	Retries              int `yaml:"retries,omitempty"`
}

func (r ConsumerIntegrationTest) SavesArtifacts() bool {
	return false
}

func (r ConsumerIntegrationTest) ReadsFromArtifacts() bool {
	return false
}

func (r ConsumerIntegrationTest) GetAttempts() int {
	return 1 + r.Retries
}

type DeployMLZip struct {
	Type          string
	Name          string
	Parallel      bool   `yaml:"parallel,omitempty"`
	DeployZip     string `json:"deploy_zip"`
	AppName       string `json:"app_name"`
	AppVersion    string `json:"app_version"`
	Targets       []string
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger"`
	Retries       int  `yaml:"retries,omitempty"`
}

func (r DeployMLZip) SavesArtifacts() bool {
	return false
}

func (r DeployMLZip) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLZip) ReadsFromArtifacts() bool {
	return true
}

type DeployMLModules struct {
	Type             string
	Name             string
	Parallel         bool   `yaml:"parallel,omitempty"`
	MLModulesVersion string `json:"ml_modules_version"`
	AppName          string `json:"app_name"`
	AppVersion       string `json:"app_version"`
	Targets          []string
	ManualTrigger    bool `json:"manual_trigger" yaml:"manual_trigger"`
	Retries          int  `yaml:"retries,omitempty"`
}

func (r DeployMLModules) SavesArtifacts() bool {
	return false
}

func (r DeployMLModules) ReadsFromArtifacts() bool {
	return false
}

func (r DeployMLModules) GetAttempts() int {
	return 1 + r.Retries
}

type Vars map[string]string
