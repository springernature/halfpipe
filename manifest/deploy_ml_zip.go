package manifest

type DeployMLZip struct {
	Type            string
	Name            string `yaml:"name,omitempty"`
	Parallel        ParallelGroup `yaml:"parallel,omitempty"`
	DeployZip       string        `json:"deploy_zip" yaml:"deploy_zip,omitempty"`
	AppName         string        `json:"app_name" yaml:"app_name,omitempty"`
	AppVersion      string        `json:"app_version" yaml:"app_version,omitempty"`
	Targets         []string      `yaml:"targets,omitempty" secretAllowed:"true" `
	ManualTrigger   bool          `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout         string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	UseBuildVersion bool          `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
}

func (r DeployMLZip) MarshalYAML() (interface{}, error) {
	r.Type = "deploy-ml-zip"
	return r, nil
}

func (r DeployMLZip) GetName() string {
	return r.Name
}

func (r DeployMLZip) GetParallelGroup() ParallelGroup {
	return r.Parallel
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
