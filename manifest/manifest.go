package manifest

type TaskList []Task

type Manifest struct {
	Team            string
	Pipeline        string
	SlackChannel    string   `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TriggerInterval string   `json:"trigger_interval" yaml:"trigger_interval,omitempty"`
	Repo            Repo     `yaml:"repo,omitempty"`
	Tasks           TaskList
	OnFailure       TaskList `json:"on_failure" yaml:"on_failure,omitempty"`
	AutoUpdate      bool     `json:"auto_update" yaml:"auto_update,omitempty"`
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

type Task interface{}

type Run struct {
	Type             string
	Name             string
	ManualTrigger    bool     `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Script           string
	Docker           Docker
	Vars             Vars     `yaml:"vars,omitempty"`
	SaveArtifacts    []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	RestoreArtifacts bool     `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Parallel         bool     `yaml:"parallel,omitempty"`
	Retries          int      `yaml:"retries,omitempty"`
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

type DockerCompose struct {
	Type             string
	Name             string
	Command          string
	ManualTrigger    bool     `json:"manual_trigger" yaml:"manual_trigger"`
	Vars             Vars
	Service          string
	SaveArtifacts    []string `json:"save_artifacts"`
	RestoreArtifacts bool     `json:"restore_artifacts" yaml:"restore_artifacts"`
	Parallel         bool     `yaml:"parallel,omitempty"`
	Retries          int      `yaml:"retries,omitempty"`
}

type DeployCF struct {
	Type           string
	Name           string
	ManualTrigger  bool     `json:"manual_trigger" yaml:"manual_trigger"`
	API            string
	Space          string
	Org            string
	Username       string
	Password       string
	Manifest       string
	TestDomain     string   `json:"test_domain" yaml:"test_domain"`
	Vars           Vars
	DeployArtifact string   `json:"deploy_artifact"`
	PrePromote     TaskList `json:"pre_promote"`
	Parallel       bool     `yaml:"parallel,omitempty"`
	Timeout        string
	Retries        int      `yaml:"retries,omitempty"`
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
	Retries              int    `yaml:"retries,omitempty"`
}

type DeployMLZip struct {
	Type          string
	Name          string
	Parallel      bool   `yaml:"parallel,omitempty"`
	DeployZip     string `json:"deploy_zip"`
	AppName       string `json:"app_name"`
	AppVersion    string `json:"app_version"`
	Targets       []string
	ManualTrigger bool   `json:"manual_trigger" yaml:"manual_trigger"`
	Retries       int    `yaml:"retries,omitempty"`
}

type DeployMLModules struct {
	Type             string
	Name             string
	Parallel         bool   `yaml:"parallel,omitempty"`
	MLModulesVersion string `json:"ml_modules_version"`
	AppName          string `json:"app_name"`
	AppVersion       string `json:"app_version"`
	Targets          []string
	ManualTrigger    bool   `json:"manual_trigger" yaml:"manual_trigger"`
	Retries          int    `yaml:"retries,omitempty"`
}

type Vars map[string]string

func (r Run) GetAttempts() int {
	return 1 + r.Retries
}

func (r DockerCompose) GetAttempts() int {
	return 1 + r.Retries
}

func (r DockerPush) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployCF) GetAttempts() int {
	return 2 + r.Retries
}

func (r ConsumerIntegrationTest) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLModules) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLZip) GetAttempts() int {
	return 1 + r.Retries
}
