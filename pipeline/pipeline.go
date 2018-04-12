package pipeline

import (
	"fmt"
	"path"
	"strings"

	"text/template"

	"bytes"

	"path/filepath"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

type Renderer interface {
	Render(manifest manifest.Manifest) atc.Config
}

type pipeline struct {
	rManifest func(string) ([]cfManifest.Application, error)
}

func NewPipeline(rManifest func(string) ([]cfManifest.Application, error)) pipeline {
	return pipeline{rManifest: rManifest}
}

const artifactsFolderName = "artifacts"
const gitDir = "git"

func (p pipeline) addSlackResource(cfg *atc.Config, man manifest.Manifest) *atc.PlanConfig {

	slackChannelSet := man.SlackChannel != ""

	if slackChannelSet {
		slackResource := p.slackResource()
		cfg.Resources = append(cfg.Resources, slackResource)

		slackResourceType := p.slackResourceType()
		cfg.ResourceTypes = append(cfg.ResourceTypes, slackResourceType)

		return &atc.PlanConfig{
			Put: slackResource.Name,
			Params: atc.Params{
				"channel":  man.SlackChannel,
				"username": "Halfpipe",
				"icon_url": "https://concourse.halfpipe.io/public/images/favicon-failed.png",
				"text":     "The pipeline `$BUILD_PIPELINE_NAME` failed at `$BUILD_JOB_NAME`. <$ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>",
			},
		}
	}
	return nil
}

func (p pipeline) initialPlan(cfg *atc.Config, man manifest.Manifest) []atc.PlanConfig {
	initialPlan := []atc.PlanConfig{{Get: gitDir, Trigger: true}}

	if man.TriggerInterval != "" {
		timerResource := p.timerResource(man.TriggerInterval)
		cfg.Resources = append(cfg.Resources, timerResource)
		initialPlan = append(initialPlan, atc.PlanConfig{Get: timerResource.Name, Trigger: true})
	}
	return initialPlan
}

func uniqueName(cfg atc.Config, name string, defaultName string) string {
	if name == "" {
		name = defaultName
	}
	return getUniqueName(name, &cfg, 0)
}

func (p pipeline) addCfResourceType(cfg *atc.Config) {
	resTypeName := "cf-resource"
	if _, exists := cfg.ResourceTypes.Lookup(resTypeName); !exists {
		cfg.ResourceTypes = append(cfg.ResourceTypes, halfpipeCfDeployResourceType(resTypeName))
	}
}

func (p pipeline) Render(man manifest.Manifest) (cfg atc.Config) {
	cfg.Resources = append(cfg.Resources, p.gitResource(man.Repo))

	initialPlan := p.initialPlan(&cfg, man)
	failurePlan := p.addSlackResource(&cfg, man)
	p.addArtifactResource(&cfg, man)

	for i, t := range man.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case manifest.Run:
			task.Name = uniqueName(cfg, task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			jobConfig = p.runJob(task, man)

		case manifest.DockerCompose:
			task.Name = uniqueName(cfg, task.Name, "docker-compose")
			jobConfig = p.dockerComposeJob(task, man)

		case manifest.DeployCF:
			p.addCfResourceType(&cfg)
			resourceName := uniqueName(cfg, deployCFResourceName(task), "")
			task.Name = uniqueName(cfg, task.Name, "deploy-cf")
			cfg.Resources = append(cfg.Resources, p.deployCFResource(task, resourceName))
			jobConfig = p.deployCFJob(task, resourceName, man)

		case manifest.DockerPush:
			resourceName := uniqueName(cfg, "Docker Registry", "")
			task.Name = uniqueName(cfg, task.Name, "docker-push")
			cfg.Resources = append(cfg.Resources, p.dockerPushResource(task, resourceName))
			jobConfig = p.dockerPushJob(task, resourceName, man)
		}

		if failurePlan != nil {
			jobConfig.Failure = failurePlan
		}

		//insert the initial plan
		jobConfig.Plan = append(initialPlan, jobConfig.Plan...)

		if i > 0 {
			// Previous job must have passed. Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, cfg.Jobs[i-1].Name)
		}
		cfg.Jobs = append(cfg.Jobs, jobConfig)
	}

	return
}

func (p pipeline) runJob(task manifest.Run, man manifest.Manifest) atc.JobConfig {
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
						Dir:  path.Join(gitDir, man.Repo.BasePath),
						Args: runScriptArgs(task.Script, pathToArtifactsDir(gitDir, man.Repo.BasePath), task.SaveArtifacts, pathToGitRef(gitDir, man.Repo.BasePath)),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
					},
				}}}}

	if len(task.SaveArtifacts) > 0 {
		jobConfig.Plan[0].TaskConfig.Outputs = []atc.TaskOutputConfig{
			{Name: artifactsFolderName},
		}

		artifactPut := atc.PlanConfig{
			Put: GenerateArtifactsFolderName(man.Team, man.Pipeline),
			Params: atc.Params{
				"folder":       artifactsFolderName,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
		}
		jobConfig.Plan = append(jobConfig.Plan, artifactPut)
	}

	return jobConfig
}

func (p pipeline) deployCFJob(task manifest.DeployCF, resourceName string, man manifest.Manifest) atc.JobConfig {
	manifestPath := path.Join(gitDir, man.Repo.BasePath, task.Manifest)

	testDomain := resolveDefaultDomain(task.API)
	vars := convertVars(task.Vars)

	appPath := path.Join(gitDir, man.Repo.BasePath)
	if len(task.DeployArtifact) > 0 {
		appPath = filepath.Join(GenerateArtifactsFolderName(man.Team, man.Pipeline), task.DeployArtifact)
	}

	cfCommand := func(commandName string) atc.PlanConfig {
		cfg := atc.PlanConfig{
			Put: resourceName,
			Params: atc.Params{
				"command":      commandName,
				"testDomain":   testDomain,
				"manifestPath": manifestPath,
				"appPath":      appPath,
			},
		}
		if len(vars) > 0 {
			cfg.Params["vars"] = vars
		}
		return cfg
	}

	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
	}

	if len(task.DeployArtifact) > 0 {
		artifactGet := atc.PlanConfig{
			Get: GenerateArtifactsFolderName(man.Team, man.Pipeline),
			Params: atc.Params{
				"folder":       artifactsFolderName,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
		}
		job.Plan = append(job.Plan, artifactGet)
	}

	job.Plan = append(job.Plan, cfCommand("halfpipe-push"))

	for _, t := range task.PrePromote {
		applications, e := p.rManifest(task.Manifest)
		if e != nil {
			panic(e)
		}
		appName := applications[0].Name
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = map[string]string{}
			}
			ppTask.Vars["TEST_ROUTE"] = appName + "-CANDIDATE." + testDomain
			job.Plan = append(job.Plan, p.runJob(ppTask, man).Plan[0])
		case manifest.DockerCompose:
			job.Plan = append(job.Plan, p.dockerComposeJob(ppTask, man).Plan[0])
		}
	}

	job.Plan = append(job.Plan, cfCommand("halfpipe-promote"))
	job.Plan = append(job.Plan, cfCommand("halfpipe-delete"))

	return job
}
func resolveDefaultDomain(targetAPI string) string {
	if strings.Contains(targetAPI, "api.dev.cf.springer-sbm.com") || strings.Contains(targetAPI, "((cloudfoundry.api-dev))") {
		return "dev.cf.private.springer.com"
	} else if strings.Contains(targetAPI, "api.live.cf.springer-sbm.com") || strings.Contains(targetAPI, "((cloudfoundry.api-live))") {
		return "live.cf.private.springer.com"
	} else if strings.Contains(targetAPI, "api.europe-west1.cf.gcp.springernature.io") || strings.Contains(targetAPI, "((cloudfoundry.api-gcp))") {
		return "apps.gcp.springernature.io"
	}

	return ""
}

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest) atc.JobConfig {
	vars := task.Vars
	if vars == nil {
		vars = make(map[string]string)
	}

	// it is really just a special run job, so let's reuse that
	vars["GCR_PRIVATE_KEY"] = "((gcr.private_key))"
	runTask := manifest.Run{
		Name:   task.Name,
		Script: dockerComposeScript(vars),
		Docker: manifest.Docker{
			Image: config.DockerComposeImage,
		},
		Vars:          vars,
		SaveArtifacts: task.SaveArtifacts,
	}
	job := p.runJob(runTask, man)
	job.Plan[0].Privileged = true
	return job
}

func (p pipeline) dockerPushJob(task manifest.DockerPush, resourceName string, man manifest.Manifest) atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"build": path.Join(gitDir, man.Repo.BasePath),
				}},
		},
	}
	if len(task.Vars) > 0 {
		job.Plan[0].Params["build_args"] = convertVars(task.Vars)
	}
	return job
}

func pathToArtifactsDir(repoName string, basePath string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath += "../"
	}

	artifactPath += artifactsFolderName
	return
}

func pathToGitRef(repoName string, basePath string) (gitRefPath string) {
	gitRefPath, _ = filepath.Rel(path.Join(repoName, basePath), path.Join(repoName, ".git", "ref"))
	return
}

func dockerComposeScript(vars map[string]string) string {
	envStrings := []string{"-e GIT_REVISION=${GIT_REVISION}"}
	for key := range vars {
		if key == "GCR_PRIVATE_KEY" {
			continue
		}
		envStrings = append(envStrings, fmt.Sprintf("-e %s=${%s}", key, key))
	}

	return fmt.Sprintf(`\source /docker-lib.sh
start_docker
docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io

docker-compose run %s app
rc=$?

docker-compose down

exit $rc
`, strings.Join(envStrings, " "))
}

func (p pipeline) addArtifactResource(cfg *atc.Config, man manifest.Manifest) {
	hasArtifacts := false
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if len(task.SaveArtifacts) > 0 {
				hasArtifacts = true
			}
		case manifest.DeployCF:
			if len(task.DeployArtifact) > 0 {
				hasArtifacts = true
			}
		}
	}

	if hasArtifacts {
		cfg.ResourceTypes = append(cfg.ResourceTypes, p.gcpResourceType())
		cfg.Resources = append(cfg.Resources, p.gcpResource(man.Team, man.Pipeline))
	}
}

func runScriptArgs(script string, artifactsPath string, saveArtifacts []string, pathToGitRef string) []string {
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}

	out := []string{
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef),
		script,
	}
	for _, artifact := range saveArtifacts {
		out = append(out, copyArtifactScript(artifactsPath, artifact))
	}
	return []string{"-ec", strings.Join(out, "\n")}
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
