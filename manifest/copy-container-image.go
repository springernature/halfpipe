package manifest

type CopyContainerImage struct {
	Type            string
	Name            string        `yaml:"name,omitempty"`
	ManualTrigger   bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications   Notifications `json:"notifications" yaml:"notifications,omitempty"`
	Timeout         string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	BuildHistory    int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`

	AwsAccessKeyID     string `yaml:"aws_access_key_id,omitempty" secretAllowed:"true"`
	AwsSecretAccessKey string `yaml:"aws_secret_access_key,omitempty" secretAllowed:"true"`
	Source             string `yaml:"source,omitempty"`
	Target             string `yaml:"target,omitempty"`
}

func (r CopyContainerImage) GetBuildHistory() int {
	return r.BuildHistory
}

func (r CopyContainerImage) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}

func (r CopyContainerImage) GetNotifications() Notifications {
	return r.Notifications
}

func (r CopyContainerImage) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r CopyContainerImage) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r CopyContainerImage) SetName(name string) Task {
	r.Name = name
	return r
}

func (r CopyContainerImage) MarshalYAML() (any, error) {
	r.Type = "copy-container-image"
	return r, nil
}

func (r CopyContainerImage) GetName() string {
	if r.Name == "" {
		return "copy-container-image"
	}
	return r.Name
}

func (r CopyContainerImage) GetTimeout() string {
	return r.Timeout
}

func (r CopyContainerImage) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}
func (r CopyContainerImage) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r CopyContainerImage) SavesArtifactsOnFailure() bool {
	return false
}

func (r CopyContainerImage) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r CopyContainerImage) SavesArtifacts() bool {
	return false
}

func (r CopyContainerImage) ReadsFromArtifacts() bool {
	return false
}

func (r CopyContainerImage) GetAttempts() int {
	return 1 + r.Retries
}

func (r CopyContainerImage) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
