package manifest

type Pack struct {
	Type             string        `json:"type,omitempty" yaml:"type,omitempty"`
	Buildpack        string        `json:"buildpack" yaml:"buildpack"`
	Path             string        `json:"path" yaml:"path"`
	ImageName        string        `json:"image_name,omitempty" yaml:"image_name,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	BuildHistory     int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	Notifications    Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Name             string        `json:"name,omitempty" yaml:"name,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	ManualTrigger    bool          `json:"manual_trigger,omitempty" yaml:"manual_trigger,omitempty"`
	RestoreArtifacts bool          `json:"restore_artifacts,omitempty" yaml:"restore_artifacts,omitempty"`
	Retries          int           `json:"retries,omitempty" yaml:"retries,omitempty"`
	Vars             Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
}

func (p Pack) GetSecrets() map[string]string {
	return findSecrets(map[string]string{})
}

func (p Pack) GetBuildHistory() int {
	return p.BuildHistory
}

func (p Pack) SetBuildHistory(buildHistory int) Task {
	p.BuildHistory = buildHistory
	return p
}

func (p Pack) GetNotifications() Notifications {
	return p.Notifications
}

func (p Pack) SetNotifications(notifications Notifications) Task {
	p.Notifications = notifications
	return p
}

func (p Pack) SetTimeout(timeout string) Task {
	p.Timeout = timeout
	return p
}

func (p Pack) SetName(name string) Task {
	p.Name = name
	return p
}

func (p Pack) MarshalYAML() (interface{}, error) {
	p.Type = "docker-push"
	return p, nil
}

func (p Pack) GetName() string {
	if p.Name == "" {
		return "docker-push"
	}
	return p.Name
}

func (p Pack) GetTimeout() string {
	return p.Timeout
}

func (p Pack) NotifiesOnSuccess() bool {
	return p.NotifyOnSuccess
}
func (p Pack) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	p.NotifyOnSuccess = notifyOnSuccess
	return p
}

func (p Pack) SavesArtifactsOnFailure() bool {
	return false
}

func (p Pack) IsManualTrigger() bool {
	return p.ManualTrigger
}

func (p Pack) SavesArtifacts() bool {
	return false
}

func (p Pack) ReadsFromArtifacts() bool {
	return p.RestoreArtifacts
}

func (p Pack) GetAttempts() int {
	return 1 + p.Retries
}

func (p Pack) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
