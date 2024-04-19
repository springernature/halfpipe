package manifest

type DeployMLZip struct {
	Type            string
	Name            string        `yaml:"name,omitempty"`
	DeployZip       string        `json:"deploy_zip" yaml:"deploy_zip,omitempty"`
	AppName         string        `json:"app_name" yaml:"app_name,omitempty"`
	AppVersion      string        `json:"app_version" yaml:"app_version,omitempty"`
	Targets         []string      `yaml:"targets,omitempty" secretAllowed:"true" `
	ManualTrigger   bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications   Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout         string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	UseBuildVersion bool          `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
	Username        string        `json:"username" yaml:"username,omitempty" secretAllowed:"true"`
	Password        string        `json:"password" yaml:"password,omitempty" secretAllowed:"true"`
	BuildHistory    int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`
}

func (r DeployMLZip) GetSecrets() map[string]string {
	return findSecrets(map[string]string{})
}

func (r DeployMLZip) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DeployMLZip) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
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
func (r DeployMLZip) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
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
