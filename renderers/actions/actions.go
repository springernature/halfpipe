package actions

import (
	"github.com/springernature/halfpipe/renderers/shared"
	"regexp"
	"strings"
	"time"

	"github.com/springernature/halfpipe/manifest"
)

const repoAccessToken = "${{ secrets.EE_REPO_ACCESS_TOKEN }}"
const slackToken = "${{ secrets.EE_SLACK_TOKEN }}"
const defaultRunner = "ee-runner"

var globalEnv = Env{
	"ARTIFACTORY_PASSWORD": "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	"ARTIFACTORY_URL":      "${{ secrets.EE_ARTIFACTORY_URL }}",
	"ARTIFACTORY_USERNAME": "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
	"BUILD_VERSION":        "${{ github.run_number }}",
	"GIT_REVISION":         "${{ github.sha }}",
	"RUNNING_IN_CI":        "true",
	"VAULT_ROLE_ID":        "${{ secrets.VAULT_ROLE_ID }}",
	"VAULT_SECRET_ID":      "${{ secrets.VAULT_SECRET_ID }}",
}

type Actions struct {
	workingDir string
}

func NewActions() Actions {
	return Actions{}
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.triggers(man.Triggers)
	if len(man.Tasks) > 0 {
		w.Env = globalEnv
		if basePath := man.Triggers.GetGitTrigger().BasePath; basePath != "" {
			w.Defaults.Run.WorkingDirectory = basePath
			a.workingDir = basePath
		}
		w.Jobs = a.jobs(man.Tasks, man, nil)
	}
	return w.asYAML()
}

type parentTask struct {
	isParallel bool
	needs      []string
}

func (a *Actions) jobs(tasks manifest.TaskList, man manifest.Manifest, parent *parentTask) (jobs Jobs) {
	appendJob := func(taskSteps Steps, task manifest.Task, needs []string) {
		steps := Steps{checkoutCode}
		if task.ReadsFromArtifacts() {
			steps = append(steps, a.restoreArtifacts()...)
		}
		steps = append(steps, taskSteps...)

		job := Job{
			Name:   task.GetName(),
			RunsOn: defaultRunner,
			Steps:  convertSecrets(steps, man.Team),
		}

		var saveArtifacts []string
		var saveArtifactsOnFailure []string
		switch task := task.(type) {
		case manifest.Run:
			saveArtifacts = task.SaveArtifacts
			saveArtifactsOnFailure = task.SaveArtifactsOnFailure
		case manifest.DockerCompose:
			saveArtifacts = task.SaveArtifacts
			saveArtifactsOnFailure = task.SaveArtifactsOnFailure
		}
		if task.SavesArtifacts() {
			job.Steps = append(job.Steps, a.saveArtifacts(saveArtifacts)...)
		}
		if task.SavesArtifactsOnFailure() {
			job.Steps = append(job.Steps, a.saveArtifactsOnFailure(saveArtifactsOnFailure)...)
		}

		if task.GetNotifications().NotificationsDefined() {
			job.Steps = append(job.Steps, notify(task.GetNotifications())...)
		}

		job.TimeoutMinutes = timeoutInMinutes(task.GetTimeout())
		job.Needs = needs
		jobs = append(jobs, Jobs{{Key: idFromName(job.Name), Value: job}}[0])
	}

	for i, t := range tasks {
		needs := idsFromNames(tasks.PreviousTaskNames(i))
		if parent != nil {
			if parent.isParallel || i == 0 {
				needs = parent.needs
			}
		}
		switch task := t.(type) {
		case manifest.DockerPush:
			appendJob(a.dockerPushSteps(task, man), task, needs)
		case manifest.Run:
			appendJob(a.runSteps(task), task, needs)
		case manifest.DockerCompose:
			appendJob(a.dockerComposeSteps(task), task, needs)
		case manifest.ConsumerIntegrationTest:
			appendJob(a.consumerIntegrationTestSteps(task, man), task, needs)
		case manifest.DeployMLModules:
			runTask := shared.ConvertDeployMLModules(task, man)
			appendJob(a.runSteps(runTask), task, needs)
		case manifest.DeployMLZip:
			runTask := shared.ConvertDeployMLZip(task, man)
			appendJob(a.runSteps(runTask), task, needs)
		case manifest.DeployCF:
			appendJob(a.deployCFSteps(task), task, needs)
		case manifest.Parallel:
			jobs = append(jobs, a.jobs(task.Tasks, man, &parentTask{isParallel: true, needs: needs})...)
		case manifest.Sequence:
			jobs = append(jobs, a.jobs(task.Tasks, man, &parentTask{isParallel: false, needs: needs})...)
		}
	}

	return jobs
}

var checkoutCode = Step{
	Name: "Checkout code",
	Uses: "actions/checkout@v2",
}

func timeoutInMinutes(timeout string) int {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return 60
	}
	return int(d.Minutes())
}

func idFromName(name string) string {
	re := regexp.MustCompile(`[^a-z_0-9\-]`)
	return re.ReplaceAllString(strings.ToLower(name), "_")
}

func idsFromNames(names []string) []string {
	for i, n := range names {
		names[i] = idFromName(n)
	}
	return names
}

func notify(notifications manifest.Notifications) (steps Steps) {
	s := func(channel string, text string) Step {
		return Step{
			Name: "Notify slack " + channel,
			Uses: "yukin01/slack-bot-action@v0.0.4",
			With: With{
				{"status", "${{ job.status }}"},
				{"oauth_token", slackToken},
				{"channel", channel},
				{"text", text},
			},
		}
	}

	for _, channel := range notifications.OnFailure {
		step := s(channel, notifications.OnFailureMessage)
		step.If = "failure()"
		step.Name += " (failure)"
		steps = append(steps, step)
	}

	for _, channel := range notifications.OnSuccess {
		step := s(channel, notifications.OnSuccessMessage)
		step.Name += " (success)"
		steps = append(steps, step)
	}

	return steps
}

func dockerLogin(image, username, password string) Steps {
	// check login step is needed
	if username == "" || strings.HasPrefix(image, "eu.gcr.io/halfpipe-io/") {
		return Steps{}
	}

	step := Step{
		Name: "Login to Docker Registry",
		Uses: "docker/login-action@v1",
		With: With{
			{"username", username},
			{"password", password},
		},
	}

	// set registry if not docker hub by counting slashes
	// docker hub format: repository:tag or user/repository:tag
	// other registries:  another.registry/user/repository:tag
	if strings.Count(image, "/") > 1 {
		registry := strings.Split(image, "/")[0]
		step.With = append(step.With, With{{"registry", registry}}...)
	}
	return Steps{step}
}
