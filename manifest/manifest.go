package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

type Vars map[string]string

type TaskList []Task

func (tl TaskList) NotifiesOnSuccess() bool {
	for _, task := range tl {
		if task.NotifiesOnSuccess() {
			return true
		}
	}
	return false
}

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

type ParallelGroup string

func (t ParallelGroup) IsSet() bool {
	return t != "" && t != "false"
}

type Task interface {
	ReadsFromArtifacts() bool
	GetAttempts() int
	SavesArtifacts() bool
	SavesArtifactsOnFailure() bool
	IsManualTrigger() bool
	NotifiesOnSuccess() bool
	GetTimeout() string
	GetParallelGroup() ParallelGroup
	GetName() string
	MarshalYAML() (interface{}, error) // To make sure type is always set when marshalling to yaml
}

type Trigger interface {
	GetTriggerName() string
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

type Manifest struct {
	Team           string         `yaml:"team,omitempty"`
	Pipeline       string         `yaml:"pipeline,omitempty"`
	SlackChannel   string         `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	CronTrigger    string         `json:"cron_trigger" yaml:"cron_trigger,omitempty"`
	Repo           Repo           `yaml:"repo,omitempty"`
	ArtifactConfig ArtifactConfig `json:"artifact_config,omitempty" yaml:"artifact_config,omitempty"`
	FeatureToggles FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	Triggers       TriggerList    `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Tasks          TaskList       `yaml:"tasks,omitempty"`
}

func (m Manifest) NotifiesOnFailure() bool {
	return m.SlackChannel != ""
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

	return
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
