package manifest

type DeployMLZip struct {
	Type            string
	Name            string        `yaml:"name,omitempty"`
	DeployZip       string        `yaml:"deploy_zip,omitempty"`
	AppName         string        `yaml:"app_name,omitempty"`
	AppVersion      string        `yaml:"app_version,omitempty"`
	Targets         []string      `yaml:"targets,omitempty" secretAllowed:"true" `
	ManualTrigger   bool          `yaml:"manual_trigger,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `yaml:"notify_on_success,omitempty"`
	Notifications   Notifications `yaml:"notifications,omitempty"`
	Timeout         string        `yaml:"timeout,omitempty"`
	UseBuildVersion bool          `yaml:"use_build_version,omitempty"`
}

func (r DeployMLZip) GetNotifications() Notifications {
	return r.Notifications
}

func (r DeployMLZip) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DeployMLZip) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DeployMLZip) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DeployMLZip) MarshalYAML() (interface{}, error) {
	r.Type = "deploy-ml-zip"
	return r, nil
}

func (r DeployMLZip) GetName() string {
	if r.Name == "" {
		return "deploy-ml-zip"
	}
	return r.Name
}

func (r DeployMLZip) GetTimeout() string {
	return r.Timeout
}

func (r DeployMLZip) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r DeployMLZip) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLZip) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DeployMLZip) SavesArtifacts() bool {
	return false
}

func (r DeployMLZip) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLZip) ReadsFromArtifacts() bool {
	return true
}
