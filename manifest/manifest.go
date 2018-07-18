package manifest

type TaskList []Task

type Manifest struct {
	Team            string
	Pipeline        string
	SlackChannel    string `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TriggerInterval string `json:"trigger_interval" yaml:"trigger_interval,omitempty"`
	Repo            Repo   `yaml:"repo,omitempty"`
	Tasks           TaskList
	OnFailure       TaskList `json:"on_failure" yaml:"on_failure,omitempty"`
	AutoUpdate      bool     `json:"auto_update" yaml:"auto_update,omitempty"`
}

type Repo struct {
	URI          string
	BasePath     string   `json:"-"` //don't auto unmarshal
	PrivateKey   string   `json:"private_key"`
	WatchedPaths []string `json:"watched_paths"`
	IgnoredPaths []string `json:"ignored_paths"`
	GitCryptKey  string   `json:"git_crypt_key"`
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
	ManualTrigger    bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Script           string
	Docker           Docker
	Vars             Vars     `yaml:"vars,omitempty"`
	SaveArtifacts    []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	RestoreArtifacts bool     `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Passed           string
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
	Passed           string
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
	Passed         string
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
	Passed           string
}

type ConsumerIntegrationTest struct {
	Type                 string
	Name                 string
	Consumer             string
	ConsumerHost         string `json:"consumer_host" yaml:"consumer_host"`
	ProviderHost         string `json:"provider_host" yaml:"provider_host"`
	Script               string
	DockerComposeService string `json:"docker_compose_service" yaml:"docker_compose_service"`
	Passed               string
}

type Vars map[string]string
