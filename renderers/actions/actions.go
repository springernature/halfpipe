package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/renderers/shared"
	"regexp"
	"strings"
	"time"

	"github.com/springernature/halfpipe/manifest"
)

const eeRunner = "ee-runner"

var globalEnv = Env{
	"ARTIFACTORY_PASSWORD": githubSecrets.ArtifactoryPassword,
	"ARTIFACTORY_URL":      githubSecrets.ArtifactoryURL,
	"ARTIFACTORY_USERNAME": githubSecrets.ArtifactoryUsername,
	"BUILD_VERSION":        "2.${{ github.run_number }}.0",
	"GIT_REVISION":         "${{ github.sha }}",
	"RUNNING_IN_CI":        "true",
	"VAULT_ROLE_ID":        githubSecrets.VaultRoleID,
	"VAULT_SECRET_ID":      githubSecrets.VaultSecretID,
}

type Actions struct {
	gitURI     string
	workingDir string
}

func NewActions(gitURI string) Actions {
	return Actions{gitURI: gitURI}
}

func (a Actions) PlatformURL(man manifest.Manifest) string {
	url := strings.Replace(a.gitURI, "git@github.com:", "https://github.com/", 1)
	url = strings.TrimSuffix(url, ".git")
	return fmt.Sprintf("%s/actions?query=workflow:%s", url, man.PipelineName())
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.triggers(man)
	w.Concurrency = "${{ github.workflow }}"
	if len(man.Tasks) > 0 {
		w.Env = globalEnv
		gitTrigger := man.Triggers.GetGitTrigger()
		if gitTrigger.BasePath != "" {
			w.Defaults.Run.WorkingDirectory = gitTrigger.BasePath
			a.workingDir = gitTrigger.BasePath
		}
		if a.workingDir == "" {
			a.workingDir = "."
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
		steps := checkoutCode(man.Triggers.GetGitTrigger())
		if task.ReadsFromArtifacts() {
			steps = append(steps, a.restoreArtifacts()...)
		}
		steps = append(steps, taskSteps...)
		if task.GetNotifications().NotificationsDefined() {
			steps = append(steps, notify(task.GetNotifications())...)
		}

		job := Job{
			Name:           task.GetName(),
			RunsOn:         eeRunner,
			Steps:          convertSecrets(steps, man.Team),
			TimeoutMinutes: timeoutInMinutes(task.GetTimeout()),
			Needs:          needs,
		}
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
		case manifest.Update:
			appendJob(a.updateSteps(task, man), task, needs)
		case manifest.DockerPush:
			appendJob(a.dockerPushSteps(task), task, needs)
		case manifest.Run:
			appendJob(a.runSteps(task), task, needs)
		case manifest.DockerCompose:
			appendJob(a.dockerComposeSteps(task, man.Team), task, needs)
		case manifest.ConsumerIntegrationTest:
			appendJob(a.consumerIntegrationTestSteps(task, man), task, needs)
		case manifest.DeployMLModules:
			runTask := shared.ConvertDeployMLModules(task, man)
			appendJob(a.runSteps(runTask), task, needs)
		case manifest.DeployMLZip:
			runTask := shared.ConvertDeployMLZip(task, man)
			appendJob(a.runSteps(runTask), task, needs)
		case manifest.DeployCF:
			appendJob(a.deployCFSteps(task, man), task, needs)
		case manifest.DeployKatee:
			appendJob(a.deployKateeSteps(task, man), task, needs)
		case manifest.Parallel:
			jobs = append(jobs, a.jobs(task.Tasks, man, &parentTask{isParallel: true, needs: needs})...)
		case manifest.Sequence:
			jobs = append(jobs, a.jobs(task.Tasks, man, &parentTask{isParallel: false, needs: needs})...)
		}
	}

	return jobs
}

func checkoutCode(gitTrigger manifest.GitTrigger) Steps {
	checkout := Step{
		Name: "Checkout code",
		Uses: "actions/checkout@v3",
		With: With{"lfs": true, "submodules": "recursive", "ssh-key": githubSecrets.GitHubPrivateKey},
	}
	if !gitTrigger.Shallow {
		checkout.With["fetch-depth"] = 0
	}
	steps := Steps{checkout}
	if gitTrigger.GitCryptKey != "" {
		steps = append(steps, Step{
			Name: "git-crypt unlock",
			Run:  "git-crypt unlock <(echo $GIT_CRYPT_KEY | base64 -d)",
			Env: Env{
				"GIT_CRYPT_KEY": gitTrigger.GitCryptKey,
			},
		})
	}
	return steps
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
				"status":      "${{ job.status }}",
				"oauth_token": githubSecrets.SlackToken,
				"channel":     channel,
				"text":        text},
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
	if username == "" || strings.HasPrefix(image, config.DockerRegistry) {
		return Steps{}
	}

	step := Step{
		Name: "Login to Docker Registry",
		Uses: "docker/login-action@v1",
		With: With{
			"username": username,
			"password": password,
		},
	}

	// set registry if not docker hub by counting slashes
	// docker hub format: repository:tag or user/repository:tag
	// other registries:  another.registry/user/repository:tag
	if strings.Count(image, "/") > 1 {
		registry := strings.Split(image, "/")[0]
		step.With["registry"] = registry
	}
	return Steps{step}
}
