package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

type Vars map[string]string

func (v Vars) SetVar(key, value string) Vars {
	if v == nil {
		return map[string]string{
			key: value,
		}
	}

	v[key] = value
	return v
}

type Notifications struct {
	OnSuccess        []string `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnSuccessMessage string   `json:"on_success_message,omitempty" yaml:"on_success_message,omitempty"`
	OnFailure        []string `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
	OnFailureMessage string   `json:"on_failure_message,omitempty" yaml:"on_failure_message,omitempty"`
}

func (n Notifications) NotificationsDefined() bool {
	return len(n.OnSuccess) > 0 || len(n.OnFailure) > 0
}

type TaskList []Task

func (tl TaskList) SavesArtifacts() bool {
	for _, task := range tl {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}

func (tl TaskList) SavesArtifactsOnFailure() bool {
	for _, task := range tl {
		if task.SavesArtifactsOnFailure() {
			return true
		}
	}
	return false
}

func (tl TaskList) UsesNotifications() bool {
	for _, task := range tl {
		switch task := task.(type) {
		case Parallel:
			if task.Tasks.UsesNotifications() {
				return true
			}
		case Sequence:
			if task.Tasks.UsesNotifications() {
				return true
			}
		default:
			if task.GetNotifications().NotificationsDefined() {
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

	GetBuildHistory() int
	SetBuildHistory(buildHistory int) Task

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

type DefaultValues struct {
	// Random default values that does not fit into triggers/tasks
	SlackToken      string `secretAllowed:"true"`
	RepoAccessToken string `secretAllowed:"true"`
}

type Manifest struct {
	Team           string         `yaml:"team,omitempty"`
	Pipeline       string         `yaml:"pipeline,omitempty"`
	SlackChannel   string         `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	ArtifactConfig ArtifactConfig `json:"artifact_config,omitempty" yaml:"artifact_config,omitempty"`
	FeatureToggles FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	Triggers       TriggerList    `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Tasks          TaskList       `yaml:"tasks,omitempty"`

	DefaultValues DefaultValues `json:"-" yaml:"-"`
}

func (m Manifest) PipelineName() (pipelineName string) {
	re := regexp.MustCompile(`[^A-Za-z0-9\-]`)
	sanitize := func(s string) string {
		return re.ReplaceAllString(strings.TrimSpace(s), "_")
	}

	pipelineName = m.Pipeline
	gitTrigger := m.Triggers.GetGitTrigger()

	if gitTrigger.Branch != "" && gitTrigger.Branch != "master" && gitTrigger.Branch != "main" {
		pipelineName = fmt.Sprintf("%s-%s", sanitize(m.Pipeline), sanitize(gitTrigger.Branch))
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
