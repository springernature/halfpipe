package manifest

type DeployMLModules struct {
	Type             string
	Name             string        `yaml:"name,omitempty"`
	MLModulesVersion string        `yaml:"ml_modules_version,omitempty"`
	AppName          string        `yaml:"app_name,omitempty"`
	AppVersion       string        `yaml:"app_version,omitempty"`
	Targets          []string      `yaml:"targets,omitempty" secretAllowed:"true"`
	ManualTrigger    bool          `yaml:"manual_trigger,omitempty"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `yaml:"notify_on_success,omitempty"`
	Notifications    Notifications `yaml:"notifications,omitempty"`
	Timeout          string        `yaml:"timeout,omitempty"`
	UseBuildVersion  bool          `yaml:"use_build_version,omitempty"`
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
