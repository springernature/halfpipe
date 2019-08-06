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
}

type Manifest struct {
	Team            string
	Pipeline        string
	SlackChannel    string         `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TriggerInterval string         `json:"trigger_interval" yaml:"trigger_interval,omitempty"`
	CronTrigger     string         `json:"cron_trigger" yaml:"cron_trigger,omitempty"`
	Repo            Repo           `yaml:"repo,omitempty"`
	ArtifactConfig  ArtifactConfig `json:"artifact_config,omitempty" yaml:"artifact_config,omitempty"`
	FeatureToggles  FeatureToggles `json:"feature_toggles,omitempty" yaml:"feature_toggles,omitempty"`
	Tasks           TaskList
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
	if m.Repo.Branch != "" && m.Repo.Branch != "master" {
		pipelineName = fmt.Sprintf("%s-%s", sanitize(m.Pipeline), sanitize(m.Repo.Branch))
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
