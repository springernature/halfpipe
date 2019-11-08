package manifest

type DockerPush struct {
	Type             string
	Name             string        `yaml:"name,omitempty"`
	ManualTrigger    bool          `yaml:"manual_trigger,omitempty"`
	Username         string        `yaml:"username,omitempty" secretAllowed:"true"`
	Password         string        `yaml:"password,omitempty" secretAllowed:"true"`
	Image            string        `yaml:"image,omitempty"`
	Vars             Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	RestoreArtifacts bool          `yaml:"restore_artifacts,omitempty"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `yaml:"notify_on_success,omitempty"`
	Notifications    Notifications `yaml:"notifications,omitempty"`
	Timeout          string        `yaml:"timeout,omitempty"`
	DockerfilePath   string        `yaml:"dockerfile_path,omitempty"`
	BuildPath        string        `yaml:"build_path,omitempty"`
}

func (r DockerPush) GetNotifications() Notifications {
	return r.Notifications
}

func (r DockerPush) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DockerPush) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DockerPush) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerPush) MarshalYAML() (interface{}, error) {
	r.Type = "docker-push"
	return r, nil
}

func (r DockerPush) GetName() string {
	if r.Name == "" {
		return "docker-push"
	}
	return r.Name
}

func (r DockerPush) GetTimeout() string {
	return r.Timeout
}

func (r DockerPush) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r DockerPush) SavesArtifactsOnFailure() bool {
	return false
}

func (r DockerPush) IsManualTrigger() bool {
	return r.ManualTrigger
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
