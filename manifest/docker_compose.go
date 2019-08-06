package manifest

type DockerCompose struct {
	Type                   string
	Name                   string
	Command                string
	ManualTrigger          bool `json:"manual_trigger" yaml:"manual_trigger"`
	Vars                   Vars `secretAllowed:"true"`
	Service                string
	ComposeFile            string        `json:"compose_file"`
	SaveArtifacts          []string      `json:"save_artifacts"`
	RestoreArtifacts       bool          `json:"restore_artifacts" yaml:"restore_artifacts"`
	SaveArtifactsOnFailure []string      `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
	Parallel               ParallelGroup `yaml:"parallel,omitempty"`
	Retries                int           `yaml:"retries,omitempty"`
	NotifyOnSuccess        bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout                string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (r DockerCompose) GetParallelGroup() ParallelGroup {
	return r.Parallel
}

func (r DockerCompose) GetTimeout() string {
	return r.Timeout
}

func (r DockerCompose) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r DockerCompose) SavesArtifactsOnFailure() bool {
	return len(r.SaveArtifactsOnFailure) > 0
}

func (r DockerCompose) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DockerCompose) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r DockerCompose) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r DockerCompose) GetAttempts() int {
	return 1 + r.Retries
}
