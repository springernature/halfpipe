package pipeline

import (
	"fmt"
	"path"
	"strings"

	"text/template"

	"bytes"

	"path/filepath"

	"sort"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

type Renderer interface {
	Render(manifest manifest.Manifest) atc.Config
}

type Pipeline struct{}

const artifactsFolderName = "artifacts"

func (p Pipeline) Render(man manifest.Manifest) (cfg atc.Config) {
	repoResource := p.gitResource(man.Repo)
	repoName := repoResource.Name
	cfg.Resources = append(cfg.Resources, repoResource)
	initialPlan := []atc.PlanConfig{{Get: repoName, Trigger: true}}

	if man.TriggerInterval != "" {
		timerResource := p.timerResource(man.TriggerInterval)
		cfg.Resources = append(cfg.Resources, timerResource)
		initialPlan = append(initialPlan, atc.PlanConfig{Get: timerResource.Name, Trigger: true})
	}

	slackChannelSet := man.SlackChannel != ""
	var slackPlanConfig *atc.PlanConfig

	if slackChannelSet {
		slackResource := p.slackResource()
		cfg.Resources = append(cfg.Resources, slackResource)

		slackResourceType := p.slackResourceType()
		cfg.ResourceTypes = append(cfg.ResourceTypes, slackResourceType)

		slackPlanConfig = &atc.PlanConfig{
			Put: slackResource.Name,
			Params: atc.Params{
				"channel":  man.SlackChannel,
				"username": "Halfpipe",
				"icon_url": "https://concourse.halfpipe.io/public/images/favicon-failed.png",
				"text": `$BUILD_PIPELINE_NAME failed. Check it out at:
http://concourse.halfpipe.io/builds/$BUILD_ID`,
			},
		}
	}

	if p.artifactsUsed(man) {
		cfg.ResourceTypes = append(cfg.ResourceTypes, p.gcpResourceType())
		cfg.Resources = append(cfg.Resources, p.gcpResource(repoName, man.Repo.BasePath))
	}

	uniqueName := func(name string, defaultName string) string {
		if name == "" {
			name = defaultName
		}
		return getUniqueName(name, &cfg, 0)
	}

	var haveCfResourceConfig bool
	for i, t := range man.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case manifest.Run:
			task.Script = fmt.Sprintf("./%s", strings.Replace(task.Script, "./", "", 1))
			task.Name = uniqueName(task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			jobConfig = p.runJob(task, repoName, man.Repo.BasePath)
		case manifest.DeployCF:
			if !haveCfResourceConfig {
				cfg.ResourceTypes = append(cfg.ResourceTypes, halfpipeCfDeployResourceType())
				haveCfResourceConfig = true
			}
			resourceName := uniqueName(deployCFResourceName(task), "")
			task.Name = uniqueName(task.Name, "deploy-cf")
			cfg.Resources = append(cfg.Resources, p.deployCFResource(task, resourceName))
			jobConfig = p.deployCFJob(task, repoName, resourceName, man.Repo.BasePath)
		case manifest.DockerPush:
			resourceName := uniqueName("Docker Registry", "")
			task.Name = uniqueName(task.Name, "docker-push")
			cfg.Resources = append(cfg.Resources, p.dockerPushResource(task, resourceName))
			jobConfig = p.dockerPushJob(task, repoName, resourceName, man.Repo.BasePath)
		case manifest.DockerCompose:
			task.Name = uniqueName(task.Name, "docker-compose")
			jobConfig = p.dockerComposeJob(task, repoName, man.Repo.BasePath)
		}

		if slackChannelSet {
			jobConfig.Failure = slackPlanConfig
		}

		//insert the initial plan
		jobConfig.Plan = append(initialPlan, jobConfig.Plan...)

		if i > 0 {
			// Previous job must have passed. Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, cfg.Jobs[i-1].Name)
		}
		cfg.Jobs = append(cfg.Jobs, jobConfig)
		p.sortGetJobsFirst(&jobConfig)
	}

	return
}

func (p Pipeline) runJob(task manifest.Run, repoName, basePath string) atc.JobConfig {
	jobConfig := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Task: "run",
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        task.Vars,
					ImageResource: p.imageResource(task.Docker),
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  path.Join(repoName, basePath),
						Args: []string{"-ec", task.Script},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: repoName},
					},
				}}}}

	if len(task.SaveArtifacts) > 0 {
		jobConfig.Plan[0].TaskConfig.Outputs = []atc.TaskOutputConfig{
			{Name: artifactsFolderName},
		}

		artifactsPath := p.pathToArtifactsDir(repoName, basePath)
		runScriptWithCopyArtifacts := runScriptWithCopyArtifacts(artifactsPath, task.Script, task.SaveArtifacts)
		runArgs := []string{"-ec", runScriptWithCopyArtifacts}
		jobConfig.Plan[0].TaskConfig.Run.Args = runArgs

		artifactPut := atc.PlanConfig{
			Put: GenerateArtifactsFolderName(repoName, basePath),
			Params: atc.Params{
				"folder":       artifactsFolderName,
				"version_file": path.Join(repoName, ".git", "ref"),
			},
		}
		jobConfig.Plan = append(jobConfig.Plan, artifactPut)
	}

	return jobConfig
}

func (p Pipeline) deployCFJob(task manifest.DeployCF, repoName, resourceName string, basePath string) atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"manifestPath": path.Join(repoName, basePath, task.Manifest),
					"appPath":      path.Join(repoName, basePath),
				},
			},
		},
	}
	if len(task.Vars) > 0 {
		job.Plan[0].Params["vars"] = convertVars(task.Vars)
	}
	if len(task.DeployArtifact) > 0 {
		job.Plan[0].Params["appPath"] = filepath.Join(GenerateArtifactsFolderName(repoName, basePath), task.DeployArtifact)

		artifactGet := atc.PlanConfig{
			Get: GenerateArtifactsFolderName(repoName, basePath),
			Params: atc.Params{
				"folder":       artifactsFolderName,
				"version_file": path.Join(repoName, ".git", "ref"),
			},
		}
		job.Plan = append(job.Plan, artifactGet)
	}
	return job
}

func (p Pipeline) dockerComposeJob(task manifest.DockerCompose, repoName, basePath string) atc.JobConfig {
	// it is really just a special run job, so let's reuse that
	runTask := manifest.Run{
		Name:   task.Name,
		Script: p.dockerComposeScript(),
		Docker: manifest.Docker{
			Image: config.DockerComposeImage,
		},
		Vars:          task.Vars,
		SaveArtifacts: task.SaveArtifacts,
	}
	job := p.runJob(runTask, repoName, basePath)
	job.Plan[0].Privileged = true
	return job
}

func (p Pipeline) dockerPushJob(task manifest.DockerPush, repoName, resourceName string, basePath string) atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"build": path.Join(repoName, basePath),
				}},
		},
	}
	if len(task.Vars) > 0 {
		job.Plan[0].Params["build_args"] = convertVars(task.Vars)
	}
	return job
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

func (Pipeline) dockerComposeScript() string {
	return `source /docker-lib.sh
start_docker
docker-compose up --force-recreate --exit-code-from app`
}

func (Pipeline) sortGetJobsFirst(job *atc.JobConfig) {
	sort.SliceStable(job.Plan, func(i, j int) bool {
		return job.Plan[i].Get != "" && job.Plan[j].Put != ""
	})
}

func (Pipeline) artifactsUsed(man manifest.Manifest) bool {
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if len(task.SaveArtifacts) > 0 {
				return true
			}
		case manifest.DeployCF:
			if len(task.DeployArtifact) > 0 {
				return true
			}
		}
	}

	return false
}

func runScriptWithCopyArtifacts(artifactsPath string, script string, saveArtifacts []string) string {
	out := []string{script}
	for _, artifact := range saveArtifacts {
		out = append(out, copyArtifactScript(artifactsPath, artifact))
	}
	return strings.Join(out, "\n")
}

func copyArtifactScript(artifactsPath string, saveArtifact string) string {
	tmpl, err := template.New("runScript").Parse(`
if [ -d {{.SaveArtifactTask}} ]
then
  mkdir -p {{.PathToArtifact}}/{{.SaveArtifactTask}}
  cp -r {{.SaveArtifactTask}}/. {{.PathToArtifact}}/{{.SaveArtifactTask}}/
elif [ -f {{.SaveArtifactTask}} ]
then
  artifactDir=$(dirname {{.SaveArtifactTask}})
  mkdir -p {{.PathToArtifact}}/$artifactDir
  cp {{.SaveArtifactTask}} {{.PathToArtifact}}/$artifactDir
else
  echo "ERROR: Artifact '{{.SaveArtifactTask}}' not found. Try fly hijack to check the filesystem."
  exit 1
fi
`)

	if err != nil {
		panic(err)
	}

	byteBuffer := new(bytes.Buffer)
	err = tmpl.Execute(byteBuffer, struct {
		PathToArtifact   string
		SaveArtifactTask string
	}{
		artifactsPath,
		saveArtifact,
	})

	if err != nil {
		panic(err)
	}

	return byteBuffer.String()
}
