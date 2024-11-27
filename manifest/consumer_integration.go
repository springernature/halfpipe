package manifest

type ConsumerIntegrationTest struct {
	Type                   string
	Name                   string        `yaml:"name,omitempty"`
	Consumer               string        `yaml:"consumer,omitempty"`
	ConsumerHost           string        `json:"consumer_host" yaml:"consumer_host,omitempty"`
	GitCloneOptions        string        `json:"git_clone_options,omitempty" yaml:"git_clone_options,omitempty"`
	ProviderHost           string        `json:"provider_host" yaml:"provider_host,omitempty"`
	ProviderName           string        `json:"provider_name" yaml:"provider_name,omitempty"`
	Script                 string        `yaml:"script,omitempty"`
	DockerComposeFile      string        `json:"docker_compose_file" yaml:"docker_compose_file,omitempty"`
	DockerComposeService   string        `json:"docker_compose_service" yaml:"docker_compose_service,omitempty"`
	Vars                   Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	Retries                int           `yaml:"retries,omitempty"`
	NotifyOnSuccess        bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications          Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout                string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	BuildHistory           int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	UseCovenant            bool          `json:"use_covenant,omitempty" yaml:"use_covenant,omitempty"`
	SaveArtifacts          []string      `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	SaveArtifactsOnFailure []string      `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
}

func (r ConsumerIntegrationTest) GetSecrets() map[string]string {
	return findSecrets(map[string]string{})
}

func (r ConsumerIntegrationTest) GetBuildHistory() int {
	return r.BuildHistory
}

func (r ConsumerIntegrationTest) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}
func (r ConsumerIntegrationTest) GetNotifications() Notifications {
	return r.Notifications
}

func (r ConsumerIntegrationTest) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r ConsumerIntegrationTest) SetName(name string) Task {
	r.Name = name
	return r
}

func (r ConsumerIntegrationTest) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r ConsumerIntegrationTest) MarshalYAML() (interface{}, error) {
	r.Type = "consumer-integration-test"
	return r, nil
}

func (r ConsumerIntegrationTest) GetName() string {
	if r.Name == "" {
		return "consumer-integration-test"
	}
	return r.Name
}

func (r ConsumerIntegrationTest) GetTimeout() string {
	return r.Timeout
}

func (r ConsumerIntegrationTest) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}
func (r ConsumerIntegrationTest) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r ConsumerIntegrationTest) SavesArtifactsOnFailure() bool {
	return len(r.SaveArtifactsOnFailure) > 0
}

func (r ConsumerIntegrationTest) IsManualTrigger() bool {
	return false
}

func (r ConsumerIntegrationTest) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r ConsumerIntegrationTest) ReadsFromArtifacts() bool {
	return false
}

func (r ConsumerIntegrationTest) GetAttempts() int {
	return 1 + r.Retries
}

func (r ConsumerIntegrationTest) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
