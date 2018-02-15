package pipeline

import (
	"fmt"
	"strings"

	"github.com/concourse/atc"
	"github.com/ghodss/yaml"
	"github.com/springernature/halfpipe/model"
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

func (pipeline Pipeline) cfDeployResource(deployCF model.DeployCF, taskIndex int) atc.ResourceConfig {
	sources := atc.Source{
		"api":          deployCF.Api,
		"organization": deployCF.Org,
		"space":        deployCF.Space,
		"username":     deployCF.Username,
		"password":     deployCF.Password,
	}

	return atc.ResourceConfig{
		Name:   fmt.Sprintf("resource-deploy-cf_Task%v", taskIndex),
		Type:   "cf",
		Source: sources,
	}
}

func (p Pipeline) makeImageResource(image string) *atc.ImageResource {
	repo, tag := image, "latest"
	if strings.Contains(image, ":") {
		split := strings.Split(image, ":")
		repo = split[0]
		tag = split[1]
	}
	return &atc.ImageResource{
		Type: "docker-image",
		Source: atc.Source{
			"repository": repo,
			"tag":        tag,
		},
	}
}

func (p Pipeline) makeRunJob(task model.Run, repo model.Repo) atc.JobConfig {
	return atc.JobConfig{
		Name:   task.Script,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repo.GetName(), Trigger: true},
			atc.PlanConfig{
				Task: task.Script,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        task.Vars,
					ImageResource: p.makeImageResource(task.Image),
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  repo.GetName(),
						Args: []string{"-exc", fmt.Sprintf("./%s", task.Script)},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: repo.GetName()},
					},
				}}}}
}

func convertVars(vars model.Vars) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range vars {
		out[k] = v
	}
	return out
}

func (p Pipeline) makeCfDeployJob(task model.DeployCF, repoName string, taskIndex int) atc.JobConfig {
	return atc.JobConfig{
		Name:   "deploy-cf",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Put: fmt.Sprintf("resource-deploy-cf_Task%v", taskIndex),
				Params: atc.Params{
					"manifest":              task.Manifest,
					"environment_variables": convertVars(task.Vars),
				},
			},
		},
	}
}

func (p Pipeline) Render(manifest model.Manifest) atc.Config {
	config := atc.Config{
		Resources: atc.ResourceConfigs{
			p.gitResource(manifest.Repo),
		},
	}

	for i, t := range manifest.Tasks {
		switch task := t.(type) {
		case model.Run:
			config.Jobs = append(config.Jobs, p.makeRunJob(task, manifest.Repo))
		case model.DeployCF:
			config.Resources = append(config.Resources, p.cfDeployResource(task, i))
			config.Jobs = append(config.Jobs, p.makeCfDeployJob(task, manifest.Repo.GetName(), i))
		}
	}
	return config
}

func ToString(pipeline atc.Config) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	return string(renderedPipeline), err
}
