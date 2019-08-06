package manifest

type ConsumerIntegrationTest struct {
	Type                 string
	Name                 string
	Consumer             string
	ConsumerHost         string `json:"consumer_host" yaml:"consumer_host"`
	GitCloneOptions      string `json:"git_clone_options,omitempty" yaml:"git_clone_options,omitempty"`
	ProviderHost         string `json:"provider_host" yaml:"provider_host"`
	Script               string
	DockerComposeService string        `json:"docker_compose_service" yaml:"docker_compose_service"`
	Parallel             ParallelGroup `yaml:"parallel,omitempty"`
	Vars                 Vars          `secretAllowed:"true"`
	Retries              int           `yaml:"retries,omitempty"`
	NotifyOnSuccess      bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout              string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (r ConsumerIntegrationTest) GetParallelGroup() ParallelGroup {
	return r.Parallel
}

func (r ConsumerIntegrationTest) GetTimeout() string {
	return r.Timeout
}

func (r ConsumerIntegrationTest) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r ConsumerIntegrationTest) SavesArtifactsOnFailure() bool {
	return false
}

func (r ConsumerIntegrationTest) IsManualTrigger() bool {
	return false
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
