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
	OnSuccess        []string `yaml:"on_success,omitempty"`
	OnSuccessMessage string   `yaml:"on_success_message,omitempty"`
	OnFailure        []string `yaml:"on_failure,omitempty"`
	OnFailureMessage string   `yaml:"on_failure_message,omitempty"`
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

type Manifest struct {
	Team           string         `yaml:"team,omitempty"`
	Pipeline       string         `yaml:"pipeline,omitempty"`
	SlackChannel   string         `yaml:"slack_channel,omitempty"`
	ArtifactConfig ArtifactConfig `yaml:"artifact_config,omitempty"`
	FeatureToggles FeatureToggles `yaml:"feature_toggles,omitempty"`
	Triggers       TriggerList    `yaml:"triggers,omitempty"`
	Tasks          TaskList       `yaml:"tasks,omitempty"`
}

func (m Manifest) PipelineName() (pipelineName string) {
	re := regexp.MustCompile(`[^A-Za-z0-9\-]`)
	sanitize := func(s string) string {
		return re.ReplaceAllString(strings.TrimSpace(s), "_")
	}

	pipelineName = m.Pipeline
	gitTrigger := m.Triggers.GetGitTrigger()

	if gitTrigger.Branch != "" && gitTrigger.Branch != "master" {
		pipelineName = fmt.Sprintf("%s-%s", sanitize(m.Pipeline), sanitize(gitTrigger.Branch))
	}

	return pipelineName
}

type ArtifactConfig struct {
	Bucket  string `yaml:"bucket,omitempty" secretAllowed:"true"`
	JSONKey string `yaml:"json_key,omitempty" secretAllowed:"true"`
}

type Repo struct {
	URI          string   `yaml:"uri,omitempty"`
	BasePath     string   `yaml:"-"` //don't auto unmarshal
	PrivateKey   string   `yaml:"private_key,omitempty" secretAllowed:"true"`
	WatchedPaths []string `yaml:"watched_paths,omitempty"`
	IgnoredPaths []string `yaml:"ignored_paths,omitempty"`
	GitCryptKey  string   `yaml:"git_crypt_key,omitempty" secretAllowed:"true"`
	Branch       string   `yaml:"branch,omitempty"`
	Shallow      bool     `yaml:"shallow,omitempty"`
}

func (repo Repo) IsPublic() bool {
	return len(repo.URI) > 4 && repo.URI[:4] == "http"
}
