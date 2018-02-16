package pipeline

import (
	"fmt"
	"strings"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"gopkg.in/yaml.v2"
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
		Name:   fmt.Sprintf("%v. Cloud Foundry", taskIndex+1),
		Type:   "cf",
		Source: sources,
	}
}

func (Pipeline) dockerResource(docker model.DockerPush, taskIndex int) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: fmt.Sprintf("%v. Docker Registry", taskIndex+1),
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Repo,
		},
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

func (p Pipeline) makeRunJob(task model.Run, repoName string, taskIndex int) atc.JobConfig {
	return atc.JobConfig{
		Name:   fmt.Sprintf("%v. Run %s", taskIndex+1, task.Script),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Task: task.Script,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        task.Vars,
					ImageResource: p.makeImageResource(task.Image),
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  repoName,
						Args: []string{"-exc", fmt.Sprintf("./%s", task.Script)},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: repoName},
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
		Name:   fmt.Sprintf("%v. deploy-cf", taskIndex+1),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Put: fmt.Sprintf("%v. deploy-cf", taskIndex+1),
				Params: atc.Params{
					"manifest":              task.Manifest,
					"environment_variables": convertVars(task.Vars),
				},
			},
		},
	}
}

func (p Pipeline) makeDockerPushJob(task model.DockerPush, repoName string, taskIndex int) atc.JobConfig {
	return atc.JobConfig{
		Name:   fmt.Sprintf("%v. docker-push", taskIndex+1),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{Put: task.GetName(), Params: atc.Params{"build": repoName}},
		},
	}
}

func (p Pipeline) Render(manifest model.Manifest) (config atc.Config) {
	config.Resources = append(config.Resources, p.gitResource(manifest.Repo))
	repoName := manifest.Repo.GetName()

	for i, t := range manifest.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case model.Run:
			jobConfig = p.makeRunJob(task, repoName, i)
		case model.DeployCF:
			config.Resources = append(config.Resources, p.cfDeployResource(task, i))
			jobConfig = p.makeCfDeployJob(task, repoName, i)
		case model.DockerPush:
			config.Resources = append(config.Resources, p.dockerResource(task, i))
			jobConfig = p.makeDockerPushJob(task, repoName, i)
		}

		if i > 0 {
			// Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, config.Jobs[i-1].Name)
		}
		config.Jobs = append(config.Jobs, jobConfig)
	}
	return
}

func ToString(pipeline atc.Config) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	return string(renderedPipeline), err
}
