package manifest

// CopyContainerImage copies an image from the halfpipe registry
// (eu.gcr.io/halfpipe-io/) to another registry. Currently only AWS ECR is
// supported as the target. Normally this would be used after a docker-push
// or buildpack task.
type CopyContainerImage struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// AWS access key ID for the target ECR registry. Defaults to shared credentials from Vault.
	AwsAccessKeyID string `json:"aws_access_key_id,omitempty" yaml:"aws_access_key_id,omitempty" secretAllowed:"true"`
	// AWS secret access key for the target ECR registry. Defaults to shared credentials from Vault.
	AwsSecretAccessKey string `json:"aws_secret_access_key,omitempty" yaml:"aws_secret_access_key,omitempty" secretAllowed:"true"`
	// Full source image URL in the halfpipe registry, with or without tag.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// Target ECR image URL or bare ECR registry URL.
	Target   string `json:"target,omitempty" yaml:"target,omitempty"`
	TaskBase `yaml:",inline"`
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

func (r CopyContainerImage) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r CopyContainerImage) SavesArtifactsOnFailure() bool {
	return false
}

func (r CopyContainerImage) SavesArtifacts() bool {
	return false
}

func (r CopyContainerImage) ReadsFromArtifacts() bool {
	return false
}
