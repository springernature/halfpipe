package manifest

// Buildpack generates a container image using Cloud Native Buildpacks and
// publishes it to the Halfpipe registry. The task uses [Paketo Buildpacks]
// which is an implementation of the Cloud Native Buildpacks specification.
//
// [Paketo Buildpacks]: https://paketo.io
type Buildpack struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Paketo builder to use. Defaults to paketobuildpacks/builder-jammy-buildpackless-base.
	Builder string `json:"builder" yaml:"builder"`
	// Buildpack identifiers to use for building the image e.g. paketo-buildpacks/java.
	Buildpacks []string `json:"buildpacks" yaml:"buildpacks"`
	// Path to the application source code to build. Defaults to current directory.
	Path string `json:"path" yaml:"path"`
	// Docker image name to build and push. Format: eu.gcr.io/halfpipe-io/<team>/<image-name>.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Task must be triggered manually (Concourse only).
	ManualTrigger bool `json:"manual_trigger,omitempty" yaml:"manual_trigger,omitempty"`
	// Restore artifacts saved by previous tasks.
	RestoreArtifacts bool `json:"restore_artifacts,omitempty" yaml:"restore_artifacts,omitempty"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Environment variables passed to the pack build command.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
}

func (p Buildpack) GetBuildHistory() int {
	return p.BuildHistory
}

func (p Buildpack) SetBuildHistory(buildHistory int) Task {
	p.BuildHistory = buildHistory
	return p
}

func (p Buildpack) GetNotifications() Notifications {
	return p.Notifications
}

func (p Buildpack) SetNotifications(notifications Notifications) Task {
	p.Notifications = notifications
	return p
}

func (p Buildpack) SetTimeout(timeout string) Task {
	p.Timeout = timeout
	return p
}

func (p Buildpack) SetName(name string) Task {
	p.Name = name
	return p
}

func (p Buildpack) MarshalYAML() (any, error) {
	p.Type = "buildpack"
	return p, nil
}

func (p Buildpack) GetName() string {
	if p.Name == "" {
		return "buildpack"
	}
	return p.Name
}

func (p Buildpack) GetTimeout() string {
	return p.Timeout
}

func (p Buildpack) NotifiesOnSuccess() bool {
	return p.NotifyOnSuccess
}
func (p Buildpack) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	p.NotifyOnSuccess = notifyOnSuccess
	return p
}

func (p Buildpack) SavesArtifactsOnFailure() bool {
	return false
}

func (p Buildpack) IsManualTrigger() bool {
	return p.ManualTrigger
}

func (p Buildpack) SavesArtifacts() bool {
	return false
}

func (p Buildpack) ReadsFromArtifacts() bool {
	return p.RestoreArtifacts
}

func (p Buildpack) GetAttempts() int {
	return 1 + p.Retries
}

func (p Buildpack) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
