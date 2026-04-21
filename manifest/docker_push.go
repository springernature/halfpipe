package manifest

type DockerPush struct {
	Type                  string        `json:"type,omitempty" yaml:"type,omitempty"`
	Name                  string        `json:"name,omitempty" yaml:"name,omitempty"`
	ManualTrigger         bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Username              string        `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	Password              string        `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	Image                 string        `json:"image,omitempty" yaml:"image,omitempty"`
	IgnoreVulnerabilities bool          `json:"ignore_vulnerabilities,omitempty" yaml:"ignore_vulnerabilities,omitempty"`
	ScanTimeout           int           `json:"scan_timeout,omitempty" yaml:"scan_timeout,omitempty"`
	Vars                  Vars          `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	Secrets               Vars          `json:"secrets,omitempty" yaml:"secrets,omitempty" secretAllowed:"true"`
	RestoreArtifacts      bool          `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Retries               int           `json:"retries,omitempty" yaml:"retries,omitempty"`
	NotifyOnSuccess       bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	Notifications         Notifications `json:"notifications" yaml:"notifications,omitempty"`
	Timeout               string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	DockerfilePath        string        `json:"dockerfile_path,omitempty" yaml:"dockerfile_path,omitempty"`
	BuildPath             string        `json:"build_path,omitempty" yaml:"build_path,omitempty"`
	Tag                   string        `json:"tag,omitempty" yaml:"tag,omitempty"`
	BuildHistory          int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	Platforms             []string      `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	UseCache              bool          `json:"use_cache,omitempty" yaml:"use_cache,omitempty"`
}

func (r DockerPush) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DockerPush) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
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

func (r DockerPush) MarshalYAML() (any, error) {
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
func (r DockerPush) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
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

func (r DockerPush) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
