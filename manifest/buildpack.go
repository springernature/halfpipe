package manifest

type Buildpack struct {
	Type             string        `json:"type,omitempty" yaml:"type,omitempty"`
	Builder          string        `json:"builder" yaml:"builder"`
	Buildpacks       []string      `json:"buildpacks" yaml:"buildpacks"`
	Path             string        `json:"path" yaml:"path"`
	Image            string        `json:"image,omitempty" yaml:"image,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Notifications    Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Name             string        `json:"name,omitempty" yaml:"name,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	ManualTrigger    bool          `json:"manual_trigger,omitempty" yaml:"manual_trigger,omitempty"`
	RestoreArtifacts bool          `json:"restore_artifacts,omitempty" yaml:"restore_artifacts,omitempty"`
	Retries          int           `json:"retries,omitempty" yaml:"retries,omitempty"`
	Vars             Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
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

func (p Buildpack) MarshalYAML() (interface{}, error) {
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
