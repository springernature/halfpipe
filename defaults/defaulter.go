package defaults

import (
	"github.com/springernature/halfpipe/model"
)

type Defaulter func(model.Manifest) model.Manifest

type Defaults struct {
	RepoPrivateKey string
	CfUsername     string
	CfPassword     string
	CfManifest     string
	CfApiAliases   map[string]string
}

func (d Defaults) Update(man model.Manifest) model.Manifest {
	if !man.Repo.IsPublic() && man.Repo.PrivateKey == "" {
		man.Repo.PrivateKey = d.RepoPrivateKey
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.DeployCF:
			if value, found := d.CfApiAliases[task.Api]; found {
				task.ApiAlias = task.Api
				task.Api = value
			}
			if task.Org == "" {
				task.Org = man.Team
			}
			if task.Username == "" {
				task.Username = d.CfUsername
			}
			if task.Password == "" {
				task.Password = d.CfPassword
			}
			if task.Manifest == "" {
				task.Manifest = d.CfManifest
			}
			man.Tasks[i] = task
		}
	}
	return man
}
