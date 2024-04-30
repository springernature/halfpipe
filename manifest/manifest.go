package manifest

import (
	"fmt"
	"golang.org/x/exp/slices"
	"regexp"
	"strings"
)

type Vars map[string]string

type Channels struct {
	OnSuccess        []string `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnSuccessMessage string   `json:"on_success_message,omitempty" yaml:"on_success_message,omitempty"`
	OnFailure        []string `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
	OnFailureMessage string   `json:"on_failure_message,omitempty" yaml:"on_failure_message,omitempty"`
}

func (t Channels) Equal(t2 Channels) bool {
	return slices.Equal(t.OnSuccess, t2.OnSuccess) &&
		t.OnSuccessMessage == t2.OnSuccessMessage &&
		slices.Equal(t.OnFailure, t2.OnFailure) &&
		t.OnFailureMessage == t2.OnFailureMessage
}

func (s Channels) NotificationsDefined() bool {
	return len(s.OnSuccess) > 0 || len(s.OnFailure) > 0
}

type NotificationChannel map[string]string

func (nc NotificationChannel) is(t string) bool {
	if _, found := nc[t]; found {
		return true
	}
	return false
}

type NotificationChannels []NotificationChannel

func (nc NotificationChannels) is(t string) (filtered NotificationChannels) {
	for _, c := range nc {
		if c.is(t) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (nc NotificationChannels) Slack() NotificationChannels {
	return nc.is("slack")
}

func (nc NotificationChannels) Teams() NotificationChannels {
	return nc.is("teams")
}

type Notifications struct {
	OnSuccess        []string             `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnSuccessMessage string               `json:"on_success_message,omitempty" yaml:"on_success_message,omitempty"`
	OnFailure        []string             `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
	OnFailureMessage string               `json:"on_failure_message,omitempty" yaml:"on_failure_message,omitempty"`
	Slack            Channels             `json:"slack,omitempty" yaml:"slack,omitempty"`
	Teams            Channels             `json:"teams,omitempty" yaml:"teams,omitempty"`
	Success          NotificationChannels `json:"success,omitempty" yaml:"success,omitempty"`
	Failure          NotificationChannels `json:"failure,omitempty" yaml:"failure,omitempty"`
}

func (n Notifications) NotificationsDefined() bool {
	return n.Slack.NotificationsDefined() || n.Teams.NotificationsDefined()
}

func (n Notifications) Equal(n2 Notifications) bool {
	return slices.Equal(n.OnFailure, n2.OnFailure) &&
		slices.Equal(n.OnSuccess, n2.OnSuccess) &&
		n.OnFailureMessage == n2.OnFailureMessage &&
		n.OnSuccessMessage == n2.OnSuccessMessage &&
		n.Slack.Equal(n2.Slack) &&
		n.Teams.Equal(n2.Teams)
}

type TaskList []Task

func (tl TaskList) SavesArtifacts() bool {
	return slices.ContainsFunc(tl, func(t Task) bool { return t.SavesArtifacts() })
}

func (tl TaskList) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(tl, func(t Task) bool { return t.SavesArtifactsOnFailure() })
}

func (tl TaskList) UsesSlackNotifications() bool {
	for _, task := range tl {
		switch task := task.(type) {
		case Parallel:
			if task.Tasks.UsesSlackNotifications() {
				return true
			}
		case Sequence:
			if task.Tasks.UsesSlackNotifications() {
				return true
			}
		default:
			if task.GetNotifications().Slack.NotificationsDefined() {
				return true
			}
		}
	}
	return false
}

func (tl TaskList) UsesTeamsNotifications() bool {
	for _, task := range tl {
		switch task := task.(type) {
		case Parallel:
			if task.Tasks.UsesTeamsNotifications() {
				return true
			}
		case Sequence:
			if task.Tasks.UsesTeamsNotifications() {
				return true
			}
		default:
			if task.GetNotifications().Teams.NotificationsDefined() {
				return true
			}
		}
	}
	return false
}

func (tl TaskList) Flatten() (updated TaskList) {
	for _, t := range tl {
		switch task := t.(type) {
		case DeployCF:
			copied := task
			copied.PrePromote = nil
			updated = append(updated, copied)
			updated = append(updated, task.PrePromote.Flatten()...)
		case Sequence:
			updated = append(updated, task.Tasks.Flatten()...)
		case Parallel:
			updated = append(updated, task.Tasks.Flatten()...)
		default:
			updated = append(updated, task)
		}
	}
	return
}

func (tl TaskList) GetTask(name string) Task {
	for _, t := range tl.Flatten() {
		if t.GetName() == name {
			return t
		}
	}
	return nil
}

func (tl TaskList) PreviousTaskNames(currentIndex int) []string {
	if currentIndex == 0 {
		return []string{}
	}
	return TaskNamesFromTask(tl[currentIndex-1])
}

func TaskNamesFromTask(t Task) (taskNames []string) {
	switch task := t.(type) {
	case Parallel:
		for _, subTask := range task.Tasks {
			taskNames = append(taskNames, TaskNamesFromTask(subTask)...)
		}
	case Sequence:
		lastTask := task.Tasks[len(task.Tasks)-1]
		taskNames = append(taskNames, TaskNamesFromTask(lastTask)...)
	default:
		taskNames = append(taskNames, task.GetName())
	}

	return taskNames
}

type Task interface {
	ReadsFromArtifacts() bool
	GetAttempts() int
	SavesArtifacts() bool
	SavesArtifactsOnFailure() bool
	IsManualTrigger() bool
	NotifiesOnSuccess() bool

	GetTimeout() string
	SetTimeout(timeout string) Task

	GetName() string
	SetName(name string) Task

	GetNotifications() Notifications
	SetNotifications(notifications Notifications) Task
	SetNotifyOnSuccess(notifyOnSuccess bool) Task

	GetBuildHistory() int
	SetBuildHistory(buildHistory int) Task

	GetSecrets() map[string]string

	MarshalYAML() (interface{}, error) // To make sure type is always set when marshalling to yaml
}

type Trigger interface {
	GetTriggerName() string
	GetTriggerAttempts() int
	MarshalYAML() (interface{}, error) // To make sure type is always set when marshalling to yaml
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

type Manifest struct {
	Team                string         `yaml:"team,omitempty"`
	Pipeline            string         `yaml:"pipeline,omitempty"`
	SlackChannel        string         `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TeamsWebhook        string         `json:"teams_webhook,omitempty" yaml:"teams_webhook,omitempty" secretAllowed:"true"`
	SlackSuccessMessage string         `json:"slack_success_message,omitempty" yaml:"slack_success_message,omitempty"`
	SlackFailureMessage string         `json:"slack_failure_message,omitempty" yaml:"slack_failure_message,omitempty"`
	ArtifactConfig      ArtifactConfig `json:"artifact_config,omitempty" yaml:"artifact_config,omitempty"`
	FeatureToggles      FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	Triggers            TriggerList    `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Tasks               TaskList       `yaml:"tasks,omitempty"`
	Platform            Platform       `json:"platform,omitempty" yaml:"platform,omitempty"`
	Notifications       Notifications  `json:"notifications,omitempty" yaml:"notifications,omitempty"`
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

type ArtifactConfig struct {
	Bucket  string `json:"bucket" yaml:"bucket,omitempty" secretAllowed:"true"`
	JSONKey string `json:"json_key" yaml:"json_key,omitempty" secretAllowed:"true"`
}

type Repo struct {
	URI          string   `json:"uri,omitempty" yaml:"uri,omitempty"`
	BasePath     string   `json:"-" yaml:"-"` //don't auto unmarshal
	PrivateKey   string   `json:"private_key,omitempty" yaml:"private_key,omitempty" secretAllowed:"true"`
	WatchedPaths []string `json:"watched_paths,omitempty" yaml:"watched_paths,omitempty"`
	IgnoredPaths []string `json:"ignored_paths,omitempty" yaml:"ignored_paths,omitempty"`
	GitCryptKey  string   `json:"git_crypt_key,omitempty" yaml:"git_crypt_key,omitempty" secretAllowed:"true"`
	Branch       string   `json:"branch,omitempty" yaml:"branch,omitempty"`
	Shallow      bool     `json:"shallow,omitempty" yaml:"shallow,omitempty"`
}

func (repo Repo) IsPublic() bool {
	return len(repo.URI) > 4 && repo.URI[:4] == "http"
}

func findSecrets(vars map[string]string) (secrets map[string]string) {
	secrets = make(map[string]string)
	for k, v := range vars {
		if regexp.MustCompile(`\(\(.*\)\)`).MatchString(v) {
			secrets[k] = v
		}
	}
	return
}
