package manifest

type DeployMLModules struct {
	Type             string
	Name             string
	Parallel         ParallelGroup `yaml:"parallel,omitempty"`
	MLModulesVersion string        `json:"ml_modules_version"`
	AppName          string        `json:"app_name"`
	AppVersion       string        `json:"app_version"`
	Targets          []string      `secretAllowed:"true"`
	ManualTrigger    bool          `json:"manual_trigger" yaml:"manual_trigger"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	UseBuildVersion  bool          `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
}

func (r DeployMLModules) GetParallelGroup() ParallelGroup {
	return r.Parallel
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