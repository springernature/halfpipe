package manifest

type DockerPush struct {
	Type             string
	Name             string
	ManualTrigger    bool   `json:"manual_trigger" yaml:"manual_trigger"`
	Username         string `secretAllowed:"true"`
	Password         string `secretAllowed:"true"`
	Image            string
	Vars             Vars          `secretAllowed:"true"`
	RestoreArtifacts bool          `json:"restore_artifacts" yaml:"restore_artifacts"`
	Parallel         ParallelGroup `yaml:"parallel,omitempty"`
	Retries          int           `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout          string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	DockerfilePath   string        `json:"dockerfile_path,omitempty" yaml:"dockerfile_path,omitempty"`
	BuildPath        string        `json:"build_path,omitempty" yaml:"build_path,omitempty"`
}

func (r DockerPush) GetParallelGroup() ParallelGroup {
	return r.Parallel
}

func (r DockerPush) GetTimeout() string {
	return r.Timeout
}

func (r DockerPush) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
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
