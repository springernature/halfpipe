package manifest

type DockerCompose struct {
	Type                   string
	Name                   string   `yaml:"name,omitempty"`
	Command                string   `yaml:"command,omitempty"`
	ManualTrigger          bool     `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Vars                   Vars     `yaml:"vars,omitempty" secretAllowed:"true"`
	Service                string   `yaml:"service,omitempty"`
	ComposeFile            string   `json:"compose_file" yaml:"compose_file,omitempty"`
	SaveArtifacts          []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	RestoreArtifacts       bool     `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	SaveArtifactsOnFailure []string `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
	Retries                int      `yaml:"retries,omitempty"`
	NotifyOnSuccess        bool     `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout                string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (r DockerCompose) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerCompose) MarshalYAML() (interface{}, error) {
	r.Type = "docker-compose"
	return r, nil
}

func (r DockerCompose) GetName() string {
	if r.Name == "" {
		return "docker-compose"
	}
	return r.Name
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
