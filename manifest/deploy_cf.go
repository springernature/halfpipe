package manifest

import (
	"slices"
	"strings"

	"code.cloudfoundry.org/cli/util/manifestparser"
)

// DeployCF deploys an app to Cloud Foundry with zero downtime.
type DeployCF struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	// Cloud Foundry API endpoint. Defaults to ((cloudfoundry.api-snpaas)).
	API string `json:"api,omitempty" yaml:"api,omitempty" secretAllowed:"true"`
	// Cloud Foundry space to deploy to.
	Space string `json:"space,omitempty" yaml:"space,omitempty" secretAllowed:"true"`
	// Cloud Foundry organisation. Defaults to the value of team.
	Org string `json:"org,omitempty" yaml:"org,omitempty" secretAllowed:"true"`
	// Cloud Foundry username. Defaults to ((cloudfoundry.username)).
	Username string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	// Cloud Foundry password. Defaults to ((cloudfoundry.password)).
	Password string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	// Path to the Cloud Foundry app manifest, relative to the halfpipe manifest. Defaults to manifest.yml.
	Manifest string `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	// Domain used when pushing the app as a candidate. Derived from the API by default.
	TestDomain string `json:"test_domain" yaml:"test_domain,omitempty" secretAllowed:"true"`
	// Environment variables injected into the CF app environment.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Path to a file or directory saved by a previous task to deploy to CF.
	DeployArtifact string `json:"deploy_artifact" yaml:"deploy_artifact,omitempty"`
	// Tasks to run after the candidate is deployed but before it is promoted to live. TEST_ROUTE is injected.
	PrePromote TaskList `json:"pre_promote" yaml:"pre_promote,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// CF CLI commands to run immediately before the candidate app is started.
	PreStart []string `json:"pre_start,omitempty" yaml:"pre_start,omitempty"`
	// Use rolling deployment instead of blue-green.
	Rolling bool `json:"rolling,omitempty" yaml:"rolling,omitempty"`
	// Stop the candidate app if deployment fails.
	StopCandidateOnFailure bool `json:"stop_candidate_on_failure,omitempty" yaml:"stop_candidate_on_failure,omitempty"`
	IsDockerPush           bool `json:"-" yaml:"-"`
	// CF CLI version to use. Allowed values: cf7, cf8. Defaults to cf7.
	CliVersion string `json:"cli_version,omitempty" yaml:"cli_version,omitempty"`
	// Docker image tag to deploy. Required when deploying a Docker image: version or gitref.
	DockerTag string `json:"docker_tag,omitempty" yaml:"docker_tag,omitempty"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	// Route to configure with SSO.
	SSORoute string `json:"sso_route,omitempty" yaml:"sso_route,omitempty"`
	// GitHub environment to associate with this deployment.
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
