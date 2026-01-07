package manifest

type DeployMLModules struct {
	Type             string
	Name             string        `yaml:"name,omitempty"`
	MLModulesVersion string        `json:"ml_modules_version" yaml:"ml_modules_version,omitempty"`
	AppName          string        `json:"app_name" yaml:"app_name,omitempty"`
	AppVersion       string        `json:"app_version" yaml:"app_version,omitempty"`
	Targets          []string      `yaml:"targets,omitempty" secretAllowed:"true"`
	ManualTrigger    bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications    Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	UseBuildVersion  bool          `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
	Username         string        `json:"username" yaml:"username,omitempty" secretAllowed:"true"`
	Password         string        `json:"password" yaml:"password,omitempty" secretAllowed:"true"`
}

func (r DeployMLModules) GetNotifications() Notifications {
	return r.Notifications
}

func (r DeployMLModules) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DeployMLModules) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DeployMLModules) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DeployMLModules) MarshalYAML() (interface{}, error) {
	r.Type = "deploy-ml-modules"
	return r, nil
}

func (r DeployMLModules) GetName() string {
	if r.Name == "" {
		return "deploy-ml-modules"
	}
	return r.Name
}

func (r DeployMLModules) GetTimeout() string {
	return r.Timeout
}

func (r DeployMLModules) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}
func (r DeployMLModules) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DeployMLModules) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLModules) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DeployMLModules) SavesArtifacts() bool {
	return false
}

func (r DeployMLModules) ReadsFromArtifacts() bool {
	return false
}

func (r DeployMLModules) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLModules) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
