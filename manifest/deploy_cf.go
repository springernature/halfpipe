package manifest

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"slices"
	"strings"
)

type DeployCF struct {
	Type              string
	Name              string            `yaml:"name,omitempty"`
	ManualTrigger     bool              `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	API               string            `yaml:"api,omitempty" secretAllowed:"true"`
	Space             string            `yaml:"space,omitempty" secretAllowed:"true"`
	Org               string            `yaml:"org,omitempty" secretAllowed:"true"`
	Username          string            `yaml:"username,omitempty" secretAllowed:"true"`
	Password          string            `yaml:"password,omitempty" secretAllowed:"true"`
	Manifest          string            `yaml:"manifest,omitempty"`
	TestDomain        string            `json:"test_domain" yaml:"test_domain,omitempty" secretAllowed:"true"`
	Vars              Vars              `yaml:"vars,omitempty" secretAllowed:"true"`
	DeployArtifact    string            `json:"deploy_artifact" yaml:"deploy_artifact,omitempty"`
	PrePromote        TaskList          `json:"pre_promote" yaml:"pre_promote,omitempty"`
	Timeout           string            `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retries           int               `yaml:"retries,omitempty"`
	NotifyOnSuccess   bool              `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Notifications     Notifications     `json:"notifications" yaml:"notifications,omitempty"`
	PreStart          []string          `json:"pre_start,omitempty" yaml:"pre_start,omitempty"`
	Rolling           bool              `yaml:"rolling,omitempty"`
	IsDockerPush      bool              `json:"-" yaml:"-"`
	CliVersion        string            `json:"cli_version,omitempty" yaml:"cli_version,omitempty"`
	DockerTag         string            `json:"docker_tag,omitempty" yaml:"docker_tag,omitempty"`
	BuildHistory      int               `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	SSORoute          string            `json:"sso_route,omitempty" yaml:"sso_route,omitempty"`
	GitHubEnvironment GitHubEnvironment `json:"github_environment" yaml:"github_environment,omitempty"`

	CfApplication manifestparser.Application `json:"-" yaml:"-"`
}

func (r DeployCF) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DeployCF) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
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

func (r DeployCF) MarshalYAML() (any, error) {
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
func (r DeployCF) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DeployCF) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(r.PrePromote, func(t Task) bool { return t.SavesArtifactsOnFailure() })
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

	return slices.ContainsFunc(r.PrePromote, func(t Task) bool { return t.ReadsFromArtifacts() })
}

func (r DeployCF) GetAttempts() int {
	return 2 + r.Retries
}

func (r DeployCF) GetGitHubEnvironment() GitHubEnvironment {
	return r.GitHubEnvironment
}
