package pipeline

import (
	"fmt"
	"path"
	"strings"

	"text/template"

	"bytes"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

type Renderer interface {
	Render(manifest manifest.Manifest) atc.Config
}

type Pipeline struct{}

const artifactsFolderName = "artifacts"

func (p Pipeline) gitResource(repo manifest.Repo) atc.ResourceConfig {
	sources := atc.Source{
		"uri": repo.Uri,
	}

	if repo.PrivateKey != "" {
		sources["private_key"] = repo.PrivateKey
	}

	if len(repo.WatchedPaths) > 0 {
		sources["paths"] = repo.WatchedPaths
	}

	if len(repo.IgnoredPaths) > 0 {
		sources["ignore_paths"] = repo.IgnoredPaths
	}

	if repo.GitCryptKey != "" {
		sources["git_crypt_key"] = repo.GitCryptKey
	}

	return atc.ResourceConfig{
		Name:   repo.GetName(),
		Type:   "git",
		Source: sources,
	}
}

func (p Pipeline) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: "slack-alert",
		Type: "slack-notification",
		Source: atc.Source{
			"url": "https://hooks.slack.com/services/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",
		},
	}
}

func (p Pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "slack-notifcation",
		Type: "docker-image",
		Source: atc.Source{
			"repository": "cfcommunity/slack-notification-resource",
			"tag":        "latest",
		},
	}
}

func (p Pipeline) timerResource(interval string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:   "timer " + interval,
		Type:   "time",
		Source: atc.Source{"interval": interval},
	}
}

func (p Pipeline) deployCFResource(deployCF manifest.DeployCF, resourceName string) atc.ResourceConfig {
	sources := atc.Source{
		"api":          deployCF.Api,
		"organization": deployCF.Org,
		"space":        deployCF.Space,
		"username":     deployCF.Username,
		"password":     deployCF.Password,
	}

	return atc.ResourceConfig{
		Name:   resourceName,
		Type:   "cf",
		Source: sources,
	}
}

func (p Pipeline) dockerPushResource(docker manifest.DockerPush, resourceName string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: resourceName,
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Image,
		},
	}
}

func (p Pipeline) imageResource(docker manifest.Docker) *atc.ImageResource {
	repo, tag := docker.Image, "latest"
	if strings.Contains(docker.Image, ":") {
		split := strings.Split(docker.Image, ":")
		repo = split[0]
		tag = split[1]
	}

	source := atc.Source{
		"repository": repo,
		"tag":        tag,
	}

	if docker.Username != "" && docker.Password != "" {
		source["username"] = docker.Username
		source["password"] = docker.Password
	}

	return &atc.ImageResource{
		Type:   "docker-image",
		Source: source,
	}
}

func (Pipeline) pathToArtifactsDir(repoName string, basePath string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath += "../"
	}

	artifactPath += artifactsFolderName
	return
}

func (p Pipeline) runJob(task manifest.Run, repoName, jobName string, basePath string) atc.JobConfig {
	script := fmt.Sprintf("./%s", strings.Replace(task.Script, "./", "", 1))

	jobConfig := atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Task:       task.Script,
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        task.Vars,
					ImageResource: p.imageResource(task.Docker),
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  path.Join(repoName, basePath),
						Args: []string{"-ec", script},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: repoName},
					},
				}}}}

	if len(task.SaveArtifacts) > 0 {
		jobConfig.Plan[0].TaskConfig.Outputs = []atc.TaskOutputConfig{
			{Name: artifactsFolderName},
		}

		scriptInput := runScriptInput{
			PathToArtifact:   p.pathToArtifactsDir(repoName, basePath),
			Script:           script,
			SaveArtifactTask: task.SaveArtifacts[0],
		}

		runScriptWithCopyArtifact := scriptInput.renderRunScriptWithCopyArtifact()
		runArgs := []string{"-ec", runScriptWithCopyArtifact}
		jobConfig.Plan[0].TaskConfig.Run.Args = runArgs
	}

	return jobConfig
}

type runScriptInput struct {
	PathToArtifact   string
	Script           string
	SaveArtifactTask string
}

func (input runScriptInput) renderRunScriptWithCopyArtifact() string {
	tmpl, err := template.New("runScript").Parse(`ARTIFACTS_DIR={{.PathToArtifact}}
{{.Script}}
if [ ! -e {{.SaveArtifactTask}} ]; then
    echo "Artifact that should be at path '{{.SaveArtifactTask}}' not found! Bailing out"
    exit -1
fi

ARTIFACT_DIR_NAME=$(dirname {{.SaveArtifactTask}})
mkdir -p $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
cp {{.SaveArtifactTask}} $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
`)

	if err != nil {
		panic(err)
	}

	byteBuffer := new(bytes.Buffer)
	err = tmpl.Execute(byteBuffer, input)

	if err != nil {
		panic(err)
	}

	return byteBuffer.String()
}

func (p Pipeline) deployCFJob(task manifest.DeployCF, repoName, jobName, resourceName string, basePath string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"manifest":              path.Join(repoName, basePath, task.Manifest),
					"environment_variables": convertVars(task.Vars),
					"path":                  path.Join(repoName, basePath),
				},
			},
		},
	}
}

func (p Pipeline) dockerPushJob(task manifest.DockerPush, repoName, jobName, resourceName string, basePath string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"build": path.Join(repoName, basePath),
				}},
		},
	}
}

func (p Pipeline) Render(man manifest.Manifest) (config atc.Config) {
	repoResource := p.gitResource(man.Repo)
	repoName := repoResource.Name
	config.Resources = append(config.Resources, repoResource)
	initialPlan := []atc.PlanConfig{{Get: repoName, Trigger: true}}

	if man.TriggerInterval != "" {
		timerResource := p.timerResource(man.TriggerInterval)
		config.Resources = append(config.Resources, timerResource)
		initialPlan = append(initialPlan, atc.PlanConfig{Get: timerResource.Name, Trigger: true})
	}

	slackChannelSet := man.SlackChannel != ""

	if slackChannelSet {
		slackResource := p.slackResource()
		config.Resources = append(config.Resources, slackResource)

		slackResourceType := p.slackResourceType()
		config.ResourceTypes = append(config.ResourceTypes, slackResourceType)
	}

	uniqueName := func(name string) string {
		return getUniqueName(name, &config, 0)
	}

	for i, t := range man.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case manifest.Run:
			jobName := uniqueName(fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			jobConfig = p.runJob(task, repoName, jobName, man.Repo.BasePath)
		case manifest.DeployCF:
			resourceName := uniqueName(deployCFResourceName(task))
			jobName := uniqueName("deploy-cf")
			config.Resources = append(config.Resources, p.deployCFResource(task, resourceName))
			jobConfig = p.deployCFJob(task, repoName, jobName, resourceName, man.Repo.BasePath)
		case manifest.DockerPush:
			resourceName := uniqueName("Docker Registry")
			jobName := uniqueName("docker-push")
			config.Resources = append(config.Resources, p.dockerPushResource(task, resourceName))
			jobConfig = p.dockerPushJob(task, repoName, jobName, resourceName, man.Repo.BasePath)
		}

		if slackChannelSet {
			jobConfig.Failure = &atc.PlanConfig{
				Put: "slack-alert",
				Params: atc.Params{
					"channel": man.SlackChannel,
					"text": `The build had a result. Check it out at:
http://concourse.halfpipe.io/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME
or at:
http://concourse.halfpipe.io/builds/$BUILD_ID`,
				},
			}
		}

		//insert the initial plan
		jobConfig.Plan = append(initialPlan, jobConfig.Plan...)

		if i > 0 {
			// Previous job must have passed. Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, config.Jobs[i-1].Name)
		}
		config.Jobs = append(config.Jobs, jobConfig)
	}
	return
}
