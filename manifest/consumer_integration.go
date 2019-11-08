package manifest

type ConsumerIntegrationTest struct {
	Type                 string
	Name                 string        `yaml:"name,omitempty"`
	Consumer             string        `yaml:"consumer,omitempty"`
	ConsumerHost         string        `yaml:"consumer_host,omitempty"`
	GitCloneOptions      string        `yaml:"git_clone_options,omitempty"`
	ProviderHost         string        `yaml:"provider_host,omitempty"`
	Script               string        `yaml:"script,omitempty"`
	DockerComposeService string        `yaml:"docker_compose_service,omitempty"`
	Vars                 Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	Retries              int           `yaml:"retries,omitempty"`
	NotifyOnSuccess      bool          `yaml:"notify_on_success,omitempty"`
	Notifications        Notifications `yaml:"notifications,omitempty"`
	Timeout              string        `yaml:"timeout,omitempty"`
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
