package manifest

type ConsumerIntegrationTest struct {
	Type                 string
	Name                 string        `yaml:"name,omitempty"`
	Consumer             string        `yaml:"consumer,omitempty"`
	ConsumerHost         string        `json:"consumer_host" yaml:"consumer_host,omitempty"`
	GitCloneOptions      string        `json:"git_clone_options,omitempty" yaml:"git_clone_options,omitempty"`
	ProviderHost         string        `json:"provider_host" yaml:"provider_host,omitempty"`
	Script               string        `yaml:"script,omitempty"`
	DockerComposeService string        `json:"docker_compose_service" yaml:"docker_compose_service,omitempty"`
	Vars                 Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	Retries              int           `yaml:"retries,omitempty"`
	NotifyOnSuccess      bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications        Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout              string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
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
