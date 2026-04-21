package manifest

type CopyContainerImage struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
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
	// AWS access key ID for the target ECR registry. Defaults to shared credentials from Vault.
	AwsAccessKeyID string `json:"aws_access_key_id,omitempty" yaml:"aws_access_key_id,omitempty" secretAllowed:"true"`
	// AWS secret access key for the target ECR registry. Defaults to shared credentials from Vault.
	AwsSecretAccessKey string `json:"aws_secret_access_key,omitempty" yaml:"aws_secret_access_key,omitempty" secretAllowed:"true"`
	// Full source image URL in the halfpipe registry, with or without tag.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// Target ECR image URL or bare ECR registry URL.
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
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
