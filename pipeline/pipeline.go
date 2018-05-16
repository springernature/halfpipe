package pipeline

import (
	"fmt"
	"strings"

	"text/template"

	"bytes"

	"path/filepath"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"sort"

	"path"

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

const dockerPushResourceName = "Docker Registry"
const dockerBuildTmpDir = "docker_build"

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

	setPassed := func(job *atc.JobConfig) {
		numberOfJobs := len(cfg.Jobs)
		if numberOfJobs > 0 {
			job.Plan[0].Passed = append(job.Plan[0].Passed, cfg.Jobs[numberOfJobs-1].Name)
		}
	}

	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			task.Name = uniqueName(&cfg, task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			job := p.runJob(task, true, man)
			job.Failure = failurePlan
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Trigger = !task.ManualTrigger
			setPassed(job)
			cfg.Jobs = append(cfg.Jobs, *job)

		case manifest.DockerCompose:
			task.Name = uniqueName(&cfg, task.Name, "docker-compose")
			job := p.dockerComposeJob(task, man)
			job.Failure = failurePlan
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Trigger = !task.ManualTrigger
			setPassed(job)
			cfg.Jobs = append(cfg.Jobs, *job)

		case manifest.DeployCF:
			p.addCfResourceType(&cfg)
			resourceName := uniqueName(&cfg, deployCFResourceName(task), "")
			task.Name = uniqueName(&cfg, task.Name, "deploy-cf")
			cfg.Resources = append(cfg.Resources, p.deployCFResource(task, resourceName))
			jobs := p.deployCFJobs(task, resourceName, man, &cfg, initialPlan, failurePlan)
			jobs[0].Plan[0].Trigger = !task.ManualTrigger
			setPassed(jobs[0])
			for _, job := range jobs {
				cfg.Jobs = append(cfg.Jobs, *job)
			}

		case manifest.DockerPush:
			resourceName := uniqueName(&cfg, dockerPushResourceName, "")
			task.Name = uniqueName(&cfg, task.Name, "docker-push")
			cfg.Resources = append(cfg.Resources, p.dockerPushResource(task, resourceName))
			job := p.dockerPushJob(task, resourceName, man)
			job.Failure = failurePlan
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Trigger = !task.ManualTrigger
			setPassed(job)
			cfg.Jobs = append(cfg.Jobs, *job)

		}
	}
	return
}

func (p pipeline) runJob(task manifest.Run, checkForBash bool, man manifest.Manifest) *atc.JobConfig {
	jobConfig := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan:   atc.PlanSequence{},
	}

	if task.RestoreArtifacts {
		jobConfig.Plan = append(jobConfig.Plan, atc.PlanConfig{
			Get: GenerateArtifactsFolderName(man.Team, man.Pipeline),
		})
	}

	runPlan := atc.PlanConfig{
		Task: "run",
		TaskConfig: &atc.TaskConfig{
			Platform:      "linux",
			Params:        task.Vars,
			ImageResource: p.imageResource(task.Docker),
			Run: atc.TaskRunConfig{
				Path: "/bin/sh",
				Dir:  path.Join(gitDir, man.Repo.BasePath),
				Args: runScriptArgs(task.Script, checkForBash, pathToArtifactsDir(gitDir, man.Repo.BasePath), task.RestoreArtifacts, task.SaveArtifacts, pathToGitRef(gitDir, man.Repo.BasePath)),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
		}}

	if task.RestoreArtifacts {
		runPlan.TaskConfig.Inputs = append(runPlan.TaskConfig.Inputs, atc.TaskInputConfig{Name: GenerateArtifactsFolderName(man.Team, man.Pipeline), Path: artifactsFolderName})
	}

	jobConfig.Plan = append(jobConfig.Plan, runPlan)

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

	return &jobConfig
}

func (p pipeline) deployCFJobs(task manifest.DeployCF, resourceName string, man manifest.Manifest, cfg *atc.Config, initialPlan atc.PlanSequence, failurePlan *atc.PlanConfig) []*atc.JobConfig {
	var jobs []*atc.JobConfig

	manifestPath := path.Join(gitDir, man.Repo.BasePath, task.Manifest)

	testDomain := resolveDefaultDomain(task.API)
	vars := convertVars(task.Vars)

	appPath := path.Join(gitDir, man.Repo.BasePath)
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(GenerateArtifactsFolderName(man.Team, man.Pipeline), task.DeployArtifact)
	}

	cfCommand := func(commandName string) atc.PlanConfig {
		cfg := atc.PlanConfig{
			Put: resourceName,
			Params: atc.Params{
				"command":      commandName,
				"testDomain":   testDomain,
				"manifestPath": manifestPath,
				"appPath":      appPath,
				"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			},
		}
		if len(vars) > 0 {
			cfg.Params["vars"] = vars
		}
		return cfg
	}

	jobs = []*atc.JobConfig{{
		Name:    task.Name,
		Serial:  true,
		Plan:    initialPlan,
		Failure: failurePlan,
	}}

	if len(task.DeployArtifact) > 0 {
		artifactGet := atc.PlanConfig{
			Get: GenerateArtifactsFolderName(man.Team, man.Pipeline),
			Params: atc.Params{
				"folder":       artifactsFolderName,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
		}
		jobs[0].Plan = append(jobs[0].Plan, artifactGet)
	}

	jobs[0].Plan = append(jobs[0].Plan, cfCommand("halfpipe-push"))

	for _, t := range task.PrePromote {
		applications, e := p.rManifest(task.Manifest)
		if e != nil {
			panic(e)
		}
		testRoute := buildTestRoute(applications[0].Name, task.Space, testDomain)
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppTask.Name = uniqueName(cfg, ppTask.Name, fmt.Sprintf("run %s", strings.Replace(ppTask.Script, "./", "", 1)))
			job := p.runJob(ppTask, true, man)
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Passed = []string{jobs[0].Name}
			job.Failure = failurePlan
			jobs = append(jobs, job)
		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppTask.Name = uniqueName(cfg, ppTask.Name, "docker-compose")
			job := p.dockerComposeJob(ppTask, man)
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Passed = []string{jobs[0].Name}
			job.Failure = failurePlan
			jobs = append(jobs, job)
		case manifest.ConsumerIntegrationTest:
			ppTask.Name = uniqueName(cfg, ppTask.Name, "consumer-integration-test")
			job := p.consumerIntegrationTestJob(ppTask, man, testRoute)
			job.Plan = append(initialPlan, job.Plan...)
			job.Plan[0].Passed = []string{jobs[0].Name}
			job.Failure = failurePlan
			jobs = append(jobs, job)
		}
	}

	if len(task.PrePromote) == 0 {
		jobs[0].Plan = append(jobs[0].Plan, cfCommand("halfpipe-promote"))
	} else {
		job := &atc.JobConfig{
			Name:    task.Name + " - promote",
			Serial:  true,
			Plan:    append(initialPlan, cfCommand("halfpipe-promote")),
			Failure: failurePlan,
		}
		for _, j := range jobs[1:] {
			job.Plan[0].Passed = append(job.Plan[0].Passed, j.Name)
		}
		jobs = append(jobs, job)
	}

	cleanup := cfCommand("halfpipe-cleanup")

	jobs[len(jobs)-1].Ensure = &cleanup //where to run this?

	return jobs
}

func buildTestRoute(appName, space, testDomain string) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s", appName, space, testDomain)
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

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest) *atc.JobConfig {
	vars := task.Vars
	if vars == nil {
		vars = make(map[string]string)
	}

	// it is really just a special run job, so let's reuse that
	vars["GCR_PRIVATE_KEY"] = "((gcr.private_key))"
	runTask := manifest.Run{
		Name:   task.Name,
		Script: dockerComposeScript(task.Service, vars, task.Command),
		Docker: manifest.Docker{
			Image: config.DockerComposeImage,
		},
		Vars:             vars,
		SaveArtifacts:    task.SaveArtifacts,
		RestoreArtifacts: task.RestoreArtifacts,
	}
	job := p.runJob(runTask, false, man)
	job.Plan[0].Privileged = true
	return job
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
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
	return &job
}

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: GenerateArtifactsFolderName(man.Team, man.Pipeline)},
			atc.PlanConfig{
				Task: "Copying git repo and artifacts to a temporary build dir",
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": "alpine",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Args: []string{"-c", strings.Join([]string{
							fmt.Sprintf("cp -r %s/. %s", gitDir, dockerBuildTmpDir),
							fmt.Sprintf("cp -r %s/. %s", artifactsFolderName, path.Join(dockerBuildTmpDir, man.Repo.BasePath)),
						}, "\n")},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
						{Name: GenerateArtifactsFolderName(man.Team, man.Pipeline), Path: artifactsFolderName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				}},

			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"build": path.Join(dockerBuildTmpDir, man.Repo.BasePath),
				}},
		},
	}
	if len(task.Vars) > 0 {
		job.Plan[2].Params["build_args"] = convertVars(task.Vars)
	}
	return &job
}

func (p pipeline) dockerPushJob(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, man)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, man)
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

func dockerComposeScript(service string, vars map[string]string, containerCommand string) string {
	envStrings := []string{"-e GIT_REVISION=${GIT_REVISION}"}
	for key := range vars {
		if key == "GCR_PRIVATE_KEY" {
			continue
		}
		envStrings = append(envStrings, fmt.Sprintf("-e %s=${%s}", key, key))
	}
	sort.Strings(envStrings)

	composeCommand := fmt.Sprintf("docker-compose run %s %s", strings.Join(envStrings, " "), service)
	if containerCommand != "" {
		composeCommand = fmt.Sprintf("%s %s", composeCommand, containerCommand)
	}

	return fmt.Sprintf(`\source /docker-lib.sh
start_docker
docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io

%s
rc=$?

docker-compose down

[ $rc -eq 0 ] || exit $rc
`, composeCommand)
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

func runScriptArgs(script string, checkForBash bool, artifactsPath string, restoreArtifacts bool, saveArtifacts []string, pathToGitRef string) []string {
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}

	var out []string

	if checkForBash {
		out = append(out, `which bash > /dev/null
if [ $? != 0 ]; then
  echo "Bash is not present in the docker image"
  echo "If you script, or any of the script your script is calling depends on bash you will get a strange error message like:"
  echo "sh: yourscript.sh: command not found"
  echo "To fix, make sure your docker image contains bash!"
  echo ""
  echo ""
fi`)
	}

	if restoreArtifacts {
		out = append(out, fmt.Sprintf("cp -r %s/. .", artifactsPath))
	}

	out = append(out,
		"set -e",
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef),
		script,
	)

	for _, artifact := range saveArtifacts {
		out = append(out, copyArtifactScript(artifactsPath, artifact))
	}
	return []string{"-c", strings.Join(out, "\n")}
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
