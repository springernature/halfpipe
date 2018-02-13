package pipeline

import (
	"github.com/springernature/halfpipe/model"
	"github.com/concourse/atc"
)

type Renderer interface {
	Render(manifest model.Manifest) atc.Config
}

type Pipeline struct{}

func (Pipeline) gitResource(repo model.Repo) atc.ResourceConfig {
	sources := atc.Source{
		"uri": repo.Uri,
	}

	if repo.PrivateKey != "" {
		sources["private_key"] = repo.PrivateKey
	}

	return atc.ResourceConfig{
		Name:   repo.GetName(),
		Type:   "git",
		Source: sources,
	}
}

func (p Pipeline) Render(manifest model.Manifest) atc.Config {
	return atc.Config{
		Resources: atc.ResourceConfigs{
			p.gitResource(manifest.Repo),
		},
	}
}
