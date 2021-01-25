package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/renderers/shared"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/springernature/halfpipe/manifest"
)

const repoAccessToken = "${{ secrets.EE_REPO_ACCESS_TOKEN }}"
const slackToken = "${{ secrets.EE_SLACK_TOKEN }}"
const defaultRunner = "ubuntu-20.04"

var globalEnv = Env{
	"ARTIFACTORY_PASSWORD": "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	"ARTIFACTORY_URL":      "${{ secrets.EE_ARTIFACTORY_URL }}",
	"ARTIFACTORY_USERNAME": "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
	"BUILD_VERSION":        "${{ github.run_number }}",
	"GCR_PRIVATE_KEY":      "${{ secrets.EE_GCR_PRIVATE_KEY }}",
	"GIT_REVISION":         "${{ github.sha }}",
	"RUNNING_IN_CI":        "true",
}

type Actions struct {
	workingDir     string
	savedArtifacts map[string]bool
}

func NewActions() Actions {
	return Actions{
		savedArtifacts: make(map[string]bool),
	}
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

func (a *Actions) secretsToStep(team string, secrets map[string]string) (step Step) {
	secs := []string{}
	for k, v := range secrets {
		raw := strings.Replace(strings.Replace(v, "((", "", -1), "))", "", -1)
		parts := strings.Split(raw, ".")
		secs = append(secs, fmt.Sprintf("springernature/%s/%s %s | %s ;", team, parts[0], parts[1], k))
	}
	sort.Strings(secs)

	return Step{
		Name: "Vault secrets",
		Id:   "secrets",
		Uses: "hashicorp/vault-action@v2.1.1",
		With: With{
			{"url", "https://vault.halfpipe.io"},
			{"method", "approle"},
			{"roleId", "${{ secrets.VAULT_ROLE_ID }}"},
			{"secretId", "${{ secrets.VAULT_SECRET_ID }}"},
			{"exportEnv", "false"},
			{"secrets", strings.Join(secs, "\n") + "\n"},
		},
	}
}

func (a *Actions) replaceEnvWithSecrets(env Env, secrets map[string]string) Env {
	for k := range secrets {
		env[k] = fmt.Sprintf("${{ steps.secrets.outputs.%s }}", k)
	}
	return env
}

func (a *Actions) jobs(tasks manifest.TaskList, man manifest.Manifest, parent *parentTask) (jobs Jobs) {
	appendJob := func(job Job, task manifest.Task, needs []string) {
		if len(task.GetSecrets()) != 0 {
			job.Steps = append([]Step{a.secretsToStep(man.Team, task.GetSecrets())}, job.Steps...)
			job.Env = a.replaceEnvWithSecrets(job.Env, task.GetSecrets())
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
			appendJob(a.dockerPushJob(task, man), task, needs)
		case manifest.Run:
			appendJob(a.runJob(task), task, needs)
		case manifest.DockerCompose:
			appendJob(a.dockerComposeJob(task), task, needs)
		case manifest.ConsumerIntegrationTest:
			appendJob(a.consumerIntegrationTestJob(task, man), task, needs)
		case manifest.DeployMLModules:
			runTask := shared.ConvertDeployMLModules(task, man)
			appendJob(a.runJob(runTask), task, needs)
		case manifest.DeployMLZip:
			runTask := shared.ConvertDeployMLZip(task, man)
			appendJob(a.runJob(runTask), task, needs)
		case manifest.DeployCF:
			appendJob(a.deployCFJob(task), task, needs)
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

func notify(notifications manifest.Notifications) []Step {
	var steps []Step

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

func dockerLogin(image, username, password string) Step {
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
	return step
}

var loginHalfpipeGCR = Step{
	Name: "Login to GCR",
	Uses: "docker/login-action@v1",
	With: With{
		{"registry", "eu.gcr.io"},
		{"username", "_json_key"},
		{"password", "${{ secrets.EE_GCR_PRIVATE_KEY }}"},
	},
}
