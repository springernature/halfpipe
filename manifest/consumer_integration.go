package manifest

type ConsumerIntegrationTest struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// GitHub repository name of the consumer, with optional sub-directory for monorepos e.g. repo-name or monorepo/dir.
	Consumer string `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	// Address of the consumer application in the same environment as the provider.
	ConsumerHost string `json:"consumer_host" yaml:"consumer_host,omitempty"`
	// Custom options for git clone of the consumer repository e.g. --depth 100.
	GitCloneOptions string `json:"git_clone_options,omitempty" yaml:"git_clone_options,omitempty"`
	// Address of the provider application to test. Defaults to the candidate route in pre_promote.
	ProviderHost string `json:"provider_host" yaml:"provider_host,omitempty"`
	// Name of the provider app, exposed as DEPENDENCY_NAME. Defaults to the pipeline name.
	ProviderName string `json:"provider_name,omitempty" yaml:"provider_name,omitempty"`
	// Consumer test script to execute.
	Script string `json:"script,omitempty" yaml:"script,omitempty"`
	// Path to the consumer docker-compose file. Defaults to docker-compose.yml.
	DockerComposeFile string `json:"docker_compose_file" yaml:"docker_compose_file,omitempty"`
	// Service name in the consumer docker-compose. Defaults to code.
	DockerComposeService string `json:"docker_compose_service" yaml:"docker_compose_service,omitempty"`
	// Environment variables available to the docker-compose service.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	// Enable Covenant contract testing support.
	UseCovenant bool `json:"use_covenant,omitempty" yaml:"use_covenant,omitempty"`
	// Paths to files or directories to save for use in subsequent tasks.
	SaveArtifacts []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	// Paths to save when the task fails, useful for test reports.
	SaveArtifactsOnFailure []string `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
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

func (r ConsumerIntegrationTest) MarshalYAML() (any, error) {
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
