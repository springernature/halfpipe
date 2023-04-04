package manifest

type DeployKatee struct {
	Type            string
	Name            string        `yaml:"name,omitempty"`
	ManualTrigger   bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Timeout         string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Vars            Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	VelaManifest    string        `json:"vela_manifest,omitempty" yaml:"vela_manifest,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications   Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Tag             string        `json:"tag,omitempty" yaml:"tag,omitempty"`
	BuildHistory    int           `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	Namespace       string        `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

func (d DeployKatee) ReadsFromArtifacts() bool {
	return false
}

func (d DeployKatee) GetAttempts() int {
	return 2 + d.Retries
}

func (d DeployKatee) SavesArtifacts() bool {
	return false
}

func (d DeployKatee) SavesArtifactsOnFailure() bool {
	return false
}

func (d DeployKatee) IsManualTrigger() bool {
	return d.ManualTrigger
}

func (d DeployKatee) NotifiesOnSuccess() bool {
	return d.NotifyOnSuccess
}

func (d DeployKatee) GetTimeout() string {
	return d.Timeout
}

func (d DeployKatee) SetTimeout(timeout string) Task {
	d.Timeout = timeout
	return d
}

func (r DeployKatee) GetName() string {
	if r.Name == "" {
		return "deploy-katee"
	}
	return r.Name
}

func (r DeployKatee) SetName(name string) Task {
	r.Name = name
	return r
}

func (d DeployKatee) GetNotifications() Notifications {
	return d.Notifications
}

func (d DeployKatee) SetNotifications(notifications Notifications) Task {
	d.Notifications = notifications
	return d
}

func (d DeployKatee) GetBuildHistory() int {
	return d.BuildHistory
}

func (d DeployKatee) SetBuildHistory(buildHistory int) Task {
	d.BuildHistory = buildHistory
	return d
}

func (d DeployKatee) GetSecrets() map[string]string {
	return findSecrets(map[string]string{})
}

func (d DeployKatee) MarshalYAML() (interface{}, error) {
	d.Type = "deploy-katee"
	return d, nil
}
