package manifest

type DockerPushAWS struct {
	Type             string
	Name             string        `yaml:"name,omitempty"`
	ManualTrigger    bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Region           string        `yaml:"region,omitempty"`
	Image            string        `yaml:"image,omitempty"`
	AccessKeyID      string        `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty" secretAllowed:"true"`
	SecretAccessKey  string        `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty" secretAllowed:"true"`
	DockerfilePath   string        `json:"dockerfile_path,omitempty" yaml:"dockerfile_path,omitempty"`
	BuildPath        string        `json:"build_path,omitempty" yaml:"build_path,omitempty"`
	Tag              string        `json:"tag,omitempty" yaml:"tag,omitempty"`
	Vars             Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	Secrets          Vars          `yaml:"secrets,omitempty" secretAllowed:"true"`
	RestoreArtifacts bool          `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications    Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (r DockerPushAWS) GetNotifications() Notifications {
	return r.Notifications
}

func (r DockerPushAWS) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DockerPushAWS) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DockerPushAWS) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerPushAWS) MarshalYAML() (interface{}, error) {
	r.Type = "docker-push-aws"
	return r, nil
}

func (r DockerPushAWS) GetName() string {
	if r.Name == "" {
		return "docker-push-aws"
	}
	return r.Name
}

func (r DockerPushAWS) GetTimeout() string {
	return r.Timeout
}

func (r DockerPushAWS) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r DockerPushAWS) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DockerPushAWS) SavesArtifactsOnFailure() bool {
	return false
}

func (r DockerPushAWS) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DockerPushAWS) SavesArtifacts() bool {
	return false
}

func (r DockerPushAWS) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r DockerPushAWS) GetAttempts() int {
	return 1 + r.Retries
}

func (r DockerPushAWS) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
