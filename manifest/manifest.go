package manifest

import (
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/filechecker"
)

type TaskList []Task

type Manifest struct {
	Team            string
	Pipeline        string
	SlackChannel    string `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	TriggerInterval string `json:"trigger_interval" yaml:"trigger_interval,omitempty"`
	Repo            Repo   `yaml:"repo,omitempty"`
	Tasks           TaskList
}

type Repo struct {
	URI          string
	BasePath     string   `json:"-"` //don't auto unmarshal
	PrivateKey   string   `json:"private_key"`
	WatchedPaths []string `json:"watched_paths"`
	IgnoredPaths []string `json:"ignored_paths"`
	GitCryptKey  string   `json:"git_crypt_key"`
}

func (repo Repo) IsPublic() bool {
	return len(repo.URI) > 4 && repo.URI[:4] == "http"
}

type Docker struct {
	Image    string
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Task interface{}

type Run struct {
	Type          string
	Name          string
	Script        string
	Docker        Docker
	Vars          Vars     `yaml:"vars,omitempty"`
	SaveArtifacts []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
}

type DockerPush struct {
	Type     string
	Name     string
	Username string
	Password string
	Image    string
	Vars     Vars
}

type DeployCF struct {
	Type           string
	Name           string
	API            string
	Space          string
	Org            string
	Username       string
	Password       string
	Manifest       string
	Vars           Vars
	DeployArtifact string   `json:"deploy_artifact"`
	PrePromote     TaskList `json:"pre_promote"`
}

type DockerCompose struct {
	Type          string
	Name          string
	Vars          Vars
	SaveArtifacts []string `json:"save_artifacts"`
}

type Vars map[string]string

type ManifestReader func(dir string, fs afero.Afero) (man Manifest, err error)

func ReadManifest(dir string, fs afero.Afero) (man Manifest, err error) {
	yaml, err := filechecker.ReadFile(fs, filepath.Join(dir, config.HalfpipeFile))
	if err != nil {
		return
	}

	man, errs := Parse(yaml)
	if len(errs) != 0 {
		err = errs[0]
	}

	return
}
