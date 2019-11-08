package manifest

import "strings"

type DeployCF struct {
	Type            string
	Name            string        `yaml:"name,omitempty"`
	ManualTrigger   bool          `yaml:"manual_trigger,omitempty"`
	API             string        `yaml:"api,omitempty" secretAllowed:"true"`
	Space           string        `yaml:"space,omitempty" secretAllowed:"true"`
	Org             string        `yaml:"org,omitempty" secretAllowed:"true"`
	Username        string        `yaml:"username,omitempty" secretAllowed:"true"`
	Password        string        `yaml:"password,omitempty" secretAllowed:"true"`
	Manifest        string        `yaml:"manifest,omitempty"`
	TestDomain      string        `yaml:"test_domain,omitempty" secretAllowed:"true"`
	Vars            Vars          `yaml:"vars,omitempty" secretAllowed:"true"`
	DeployArtifact  string        `yaml:"deploy_artifact,omitempty"`
	PrePromote      TaskList      `yaml:"pre_promote,omitempty"`
	Timeout         string        `yaml:"timeout,omitempty"`
	Retries         int           `yaml:"retries,omitempty"`
	NotifyOnSuccess bool          `yaml:"notify_on_success,omitempty"`
	Notifications   Notifications `yaml:"notifications,omitempty"`
	PreStart        []string      `yaml:"pre_start,omitempty"`
}

func (r DeployCF) GetNotifications() Notifications {
	return r.Notifications
}

func (r DeployCF) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DeployCF) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DeployCF) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DeployCF) MarshalYAML() (interface{}, error) {
	r.Type = "deploy-cf"
	return r, nil
}

func (r DeployCF) GetName() string {
	if r.Name == "" {
		return "deploy-cf"
	}
	return r.Name
}

func (r DeployCF) GetTimeout() string {
	return r.Timeout
}

func (r DeployCF) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
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
