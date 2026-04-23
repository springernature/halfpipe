package manifest

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// vars are key-value pairs of environment variables. Values are coerced to strings.
type Vars map[string]string

// notification channel defines where to send a notification.
type NotificationChannel struct {
	// Microsoft Teams channel webhook URL.
	Teams string `json:"teams,omitempty" yaml:"teams,omitempty"`
	// Optional message to include in the notification.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	// Deprecated: Slack notifications are no longer supported.
	Slack string `json:"slack,omitempty" yaml:"slack,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
}

type NotificationChannels []NotificationChannel

func (nc NotificationChannels) Slack() (ncs NotificationChannels) {
	for _, n := range nc {
		if n.Slack != "" && n.Teams == "" {
			ncs = append(ncs, n)
		}
	}
	return ncs
}

func (nc NotificationChannels) Teams() (ncs NotificationChannels) {
	for _, n := range nc {
		if n.Slack == "" && n.Teams != "" {
			ncs = append(ncs, n)
		}
	}
	return ncs
}

// notifications configure which channels to notify on task success or failure.
type Notifications struct {
	// Notification channels to notify on task success.
	Success NotificationChannels `json:"success,omitempty" yaml:"success,omitempty"`
	// Notification channels to notify on task failure.
	Failure NotificationChannels `json:"failure,omitempty" yaml:"failure,omitempty"`
	// Deprecated: Slack notifications are no longer supported.
	OnSuccess []string `json:"on_success,omitempty" yaml:"on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// Deprecated: Slack notifications are no longer supported.
	OnSuccessMessage string `json:"on_success_message,omitempty" yaml:"on_success_message,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// Deprecated: Slack notifications are no longer supported.
	OnFailure []string `json:"on_failure,omitempty" yaml:"on_failure,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// Deprecated: Slack notifications are no longer supported.
	OnFailureMessage string `json:"on_failure_message,omitempty" yaml:"on_failure_message,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
}

func (n Notifications) NotificationsDefined() bool {
	return len(n.Failure) > 0 || len(n.Success) > 0
}

func (n Notifications) Equal(n2 Notifications) bool {

	return slices.Equal(n.OnFailure, n2.OnFailure) &&
		slices.Equal(n.OnSuccess, n2.OnSuccess) &&
		slices.Equal(n.Failure, n2.Failure) &&
		slices.Equal(n.Success, n2.Success) &&
		n.OnFailureMessage == n2.OnFailureMessage &&
		n.OnSuccessMessage == n2.OnSuccessMessage
}

type Trigger interface {
	GetTriggerName() string
	GetTriggerAttempts() int
	MarshalYAML() (any, error) // To make sure type is always set when marshalling to yaml
}

type TriggerList []Trigger

func (t TriggerList) GetGitTrigger() GitTrigger {
	for _, trigger := range t {
		switch trigger := trigger.(type) {
		case GitTrigger:
			return trigger
		}
	}
	return GitTrigger{}
}

func (t TriggerList) HasGitTrigger() bool {
	for _, trigger := range t {
		switch trigger.(type) {
		case GitTrigger:
			return true
		}
	}
	return false
}

type Platform string

func (p Platform) IsActions() bool {
	return p == "actions"
}

func (p Platform) IsConcourse() bool {
	return !p.IsActions()
}

type OpsLevel struct {
	Name         string `yaml:"name"`
	System       string `yaml:"system"`
	RelativePath string `yaml:"-"`
	ParseError   string `yaml:"-"`
}

type Manifest struct {
	// The platform team that owns this pipeline.
	Team string `json:"team,omitempty" yaml:"team,omitempty" jsonschema:"required"`
	// The name of the pipeline.
	Pipeline string `json:"pipeline,omitempty" yaml:"pipeline,omitempty" jsonschema:"required"`
	// The CI platform to target. Defaults to concourse.
	Platform Platform `json:"platform,omitempty" yaml:"platform,omitempty"`
	// Deprecated: Slack notifications are no longer supported.
	SlackChannel string `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// A Microsoft Teams webhook URL for pipeline-level notifications.
	TeamsWebhook string `json:"teams_webhook,omitempty" yaml:"teams_webhook,omitempty" secretAllowed:"true"`
	// Deprecated: Slack notifications are no longer supported.
	SlackSuccessMessage string `json:"slack_success_message,omitempty" yaml:"slack_success_message,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// Deprecated: Slack notifications are no longer supported.
	SlackFailureMessage string `json:"slack_failure_message,omitempty" yaml:"slack_failure_message,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=Slack notifications are no longer supported"`
	// Enable optional pipeline features
	FeatureToggles FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	// The triggers that cause this pipeline to run. Defaults to git.
	Triggers TriggerList `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	// The tasks that make up this pipeline.
	Tasks TaskList `json:"tasks,omitempty" yaml:"tasks,omitempty" jsonschema:"required"`
	// Default notifications for all tasks.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	OpsLevel      OpsLevel      `json:"-" yaml:"-"`
}

func (m Manifest) PipelineName() (pipelineName string) {
	re := regexp.MustCompile(`[^A-Za-z0-9\-]`)
	sanitize := func(s string) string {
		return re.ReplaceAllString(strings.TrimSpace(s), "_")
	}

	pipelineName = m.Pipeline

	if m.Platform.IsConcourse() {
		gitTrigger := m.Triggers.GetGitTrigger()

		if gitTrigger.Branch != "" && gitTrigger.Branch != "master" && gitTrigger.Branch != "main" {
			pipelineName = fmt.Sprintf("%s-%s", sanitize(m.Pipeline), sanitize(gitTrigger.Branch))
		}
	}

	return pipelineName
}

// GitHub environment to associate with this deployment.
type GitHubEnvironment struct {
	// Name of the GitHub environment.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// URL associated with the GitHub environment.
	Url string `json:"url,omitempty" yaml:"url,omitempty"`
}

func (g GitHubEnvironment) IsValid() bool {
	return g.Name != "" && g.Url != ""
}

type GitHubEnvironmentTask interface {
	GetGitHubEnvironment() GitHubEnvironment
}
