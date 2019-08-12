package manifest

import "strings"

type DeployCF struct {
	Type            string
	Name            string
	ManualTrigger   bool   `json:"manual_trigger" yaml:"manual_trigger"`
	API             string `secretAllowed:"true"`
	Space           string `secretAllowed:"true"`
	Org             string `secretAllowed:"true"`
	Username        string `secretAllowed:"true"`
	Password        string `secretAllowed:"true"`
	Manifest        string
	TestDomain      string        `json:"test_domain" yaml:"test_domain" secretAllowed:"true"`
	Vars            Vars          `secretAllowed:"true"`
	DeployArtifact  string        `json:"deploy_artifact"`
	PrePromote      TaskList      `json:"pre_promote"`
	Parallel        ParallelGroup `yaml:"parallel,omitempty"`
	Timeout         string        `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
}

func (r DeployCF) GetName() string {
	return r.Name
}

func (r DeployCF) GetParallelGroup() ParallelGroup {
	return r.Parallel
}

func (r DeployCF) GetTimeout() string {
	return r.Timeout
}

func (r DeployCF) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess || r.PrePromote.NotifiesOnSuccess()
}

func (r DeployCF) SavesArtifactsOnFailure() bool {
	for _, task := range r.PrePromote {
		if task.SavesArtifactsOnFailure() {
			return true
		}
	}
	return false
}

func (r DeployCF) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DeployCF) SavesArtifacts() bool {
	return false
}

func (r DeployCF) ReadsFromArtifacts() bool {
	if r.DeployArtifact != "" || strings.HasPrefix(r.Manifest, "../artifacts/") {
		return true
	}

	for _, pp := range r.PrePromote {
		if pp.ReadsFromArtifacts() {
			return true
		}
	}
	return false
}

func (r DeployCF) GetAttempts() int {
	return 2 + r.Retries
}
