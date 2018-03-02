package pipeline

import (
	"fmt"
	"path"
	"strings"

	"text/template"

	"bytes"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/parser"
)

type Renderer interface {
	Render(manifest parser.Manifest) atc.Config
}

type Pipeline struct{}

const artifactsFolderName = "artifacts"

func (p Pipeline) gitResource(repo parser.Repo) atc.ResourceConfig {
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

func (p Pipeline) deployCFResource(deployCF parser.DeployCF, resourceName string) atc.ResourceConfig {
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

func (p Pipeline) dockerPushResource(docker parser.DockerPush, resourceName string) atc.ResourceConfig {
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

func (p Pipeline) imageResource(docker parser.Docker) *atc.ImageResource {
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

func (p Pipeline) runJob(task parser.Run, repoName, jobName string, basePath string) atc.JobConfig {
	script := fmt.Sprintf("./%s", strings.Replace(task.Script, "./", "", 1))

	jobConfig := atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Task: task.Script,
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
		jobConfig.Plan[1].TaskConfig.Outputs = []atc.TaskOutputConfig{
			{Name: artifactsFolderName},
		}

		scriptInput := runScriptInput{
			PathToArtifact:   p.pathToArtifactsDir(repoName, basePath),
			Script:           script,
			SaveArtifactTask: task.SaveArtifacts[0],
		}

		runScriptWithCopyArtifact := scriptInput.renderRunScriptWithCopyArtifact()
		runArgs := []string{"-ec", runScriptWithCopyArtifact}
		jobConfig.Plan[1].TaskConfig.Run.Args = runArgs
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

func (p Pipeline) deployCFJob(task parser.DeployCF, repoName, jobName, resourceName string, basePath string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
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

func (p Pipeline) dockerPushJob(task parser.DockerPush, repoName, jobName, resourceName string, basePath string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"build": path.Join(repoName, basePath),
				}},
		},
	}
}

func (p Pipeline) Render(manifest parser.Manifest) (config atc.Config) {
	config.Resources = append(config.Resources, p.gitResource(manifest.Repo))
	repoName := manifest.Repo.GetName()

	uniqueName := func(name string) string {
		return getUniqueName(name, &config, 0)
	}

	for i, t := range manifest.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case parser.Run:
			jobName := uniqueName(fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			jobConfig = p.runJob(task, repoName, jobName, manifest.Repo.BasePath)
		case parser.DeployCF:
			resourceName := uniqueName(deployCFResourceName(task))
			jobName := uniqueName("deploy-cf")
			config.Resources = append(config.Resources, p.deployCFResource(task, resourceName))
			jobConfig = p.deployCFJob(task, repoName, jobName, resourceName, manifest.Repo.BasePath)
		case parser.DockerPush:
			resourceName := uniqueName("Docker Registry")
			jobName := uniqueName("docker-push")
			config.Resources = append(config.Resources, p.dockerPushResource(task, resourceName))
			jobConfig = p.dockerPushJob(task, repoName, jobName, resourceName, manifest.Repo.BasePath)
		}

		if i > 0 {
			// Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, config.Jobs[i-1].Name)
		}
		config.Jobs = append(config.Jobs, jobConfig)
	}
	return
}
