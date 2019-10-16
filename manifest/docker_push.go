package manifest

type DockerPush struct {
	Type             string
	Name             string `yaml:"name,omitempty"`
	ManualTrigger    bool   `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Username         string `yaml:"username,omitempty" secretAllowed:"true"`
	Password         string `yaml:"password,omitempty" secretAllowed:"true"`
	Image            string `yaml:"image,omitempty"`
	Vars             Vars   `yaml:"vars,omitempty" secretAllowed:"true"`
	RestoreArtifacts bool   `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	Retries          int    `yaml:"retries,omitempty"`
	NotifyOnSuccess  bool   `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout          string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	DockerfilePath   string `json:"dockerfile_path,omitempty" yaml:"dockerfile_path,omitempty"`
	BuildPath        string `json:"build_path,omitempty" yaml:"build_path,omitempty"`
}

func (r DockerPush) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DockerPush) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerPush) MarshalYAML() (interface{}, error) {
	r.Type = "docker-push"
	return r, nil
}

func (r DockerPush) GetName() string {
	if r.Name == "" {
		return "docker-push"
	}
	return r.Name
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
