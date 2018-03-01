package model

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/springernature/halfpipe/errors"
)

type Manifest struct {
	Team  string
	Repo  Repo
	Tasks []Task `json:"-"` //don't auto unmarshal
}

type Repo struct {
	Uri          string
	PrivateKey   string   `json:"private_key"`
	WatchedPaths []string `json:"watched_paths"`
	IgnoredPaths []string `json:"ignored_paths"`
	GitCryptKey  string   `json:"git_crypt_key"`
}

type GitPaths struct {
	Watch  []string `json:"watch"`
	Ignore []string `json:"ignore"`
}

func (repo Repo) GetName() string {
	re := regexp.MustCompile(`^(?:.+\/)([^.]+)(?:\.git\/?)?$`)
	matches := re.FindStringSubmatch(repo.Uri)
	if len(matches) != 2 {
		return repo.Uri
	}
	return matches[1]
}

func (repo Repo) IsPublic() bool {
	return len(repo.Uri) > 4 && repo.Uri[:4] == "http"
}

type Task interface{}

type Run struct {
	Script        string
	Docker        Docker
	Vars          Vars
	SaveArtifacts []string `json:"save_artifacts"`
}

type Docker struct {
	Image    string
	Username string
	Password string
}

type DockerPush struct {
	Username string
	Password string
	Image    string
	Vars     Vars
}

type DeployCF struct {
	Api            string
	Space          string
	Org            string
	Username       string
	Password       string
	Manifest       string
	Vars           Vars
	DeployArtifact string `json:"deploy_artifact"`
}

type Vars map[string]string

// convert bools and floats into strings, anything else is invalid
func (r *Vars) UnmarshalJSON(b []byte) error {
	rawVars := make(map[string]interface{})
	if err := json.Unmarshal(b, &rawVars); err != nil {
		errors.NewInvalidField("var", err.Error())
		return err
	}
	stringVars := make(Vars)

	for key, v := range rawVars {
		switch value := v.(type) {
		case string, bool, float64:
			stringVars[key] = fmt.Sprintf("%v", value)
		default:
			return errors.NewInvalidField("var", fmt.Sprintf("value of key '%v' must be a string", key))
		}
	}
	*r = stringVars
	return nil
}
