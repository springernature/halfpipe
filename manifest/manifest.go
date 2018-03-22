package manifest

import (
	"regexp"
)

type Manifest struct {
	Team            string
	TriggerInterval string `json:"trigger_interval"`
	Repo            Repo
	SlackChannel    string `json:"slack_channel"`
	Tasks           []Task
}

type Repo struct {
	URI          string
	BasePath     string   `json:"-"` //don't auto unmarshal
	PrivateKey   string   `json:"private_key"`
	WatchedPaths []string `json:"watched_paths"`
	IgnoredPaths []string `json:"ignored_paths"`
	GitCryptKey  string   `json:"git_crypt_key"`
}

func (repo Repo) GetName() string {
	re := regexp.MustCompile(`^(?:.+\/)([^.]+)(?:\.git\/?)?$`)
	matches := re.FindStringSubmatch(repo.URI)
	if len(matches) != 2 {
		return repo.URI
	}
	return matches[1]
}

func (repo Repo) IsPublic() bool {
	return len(repo.URI) > 4 && repo.URI[:4] == "http"
}

type Docker struct {
	Image    string
	Username string
	Password string
}

type Task interface{}

type Run struct {
	Type          string
	Name          string
	Script        string
	Docker        Docker
	Vars          Vars
	SaveArtifacts []string `json:"save_artifacts"`
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
	DeployArtifact string `json:"deploy_artifact"`
	PrePromote     []Task `json:"pre_promote"`
}

type DockerCompose struct {
	Type          string
	Name          string
	Vars          Vars
	SaveArtifacts []string `json:"save_artifacts"`
}

type Vars map[string]string
