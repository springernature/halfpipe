package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/springernature/halfpipe/errors"
)

type Manifest struct {
	Team  string
	Repo  Repo
	Tasks []Task `json:"-"` //don't auto unmarshal
}

type Repo struct {
	Uri        string
	PrivateKey string `json:"private_key"`
}

func (repo Repo) GetName() string {
	var withoutPostfix string
	if strings.HasSuffix(repo.Uri, ".git/") {
		withoutPostfix = strings.Split(repo.Uri, ".git/")[0]
	} else if strings.HasSuffix(repo.Uri, ".git") {
		withoutPostfix = strings.Split(repo.Uri, ".git")[0]
	} else {
		withoutPostfix = repo.Uri
	}

	parts := strings.Split(withoutPostfix, "/")
	return parts[len(parts)-1]
}

type Task interface {
	GetName() string
}

type Run struct {
	Name   string
	Script string
	Image  string
	Vars   Vars
}

func (t Run) GetName() string {
	return t.Script
}

type DockerPush struct {
	Name     string
	Username string
	Password string
	Repo     string
	Vars     Vars
}

func (t DockerPush) GetName() string {
	return "docker-push"
}

type DeployCF struct {
	Name     string
	Api      string
	Space    string
	Org      string
	Username string
	Password string
	Manifest string
	Vars     Vars
}

func (t DeployCF) GetName() string {
	return t.Name
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
