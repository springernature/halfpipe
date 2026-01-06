package defaults

import "github.com/springernature/halfpipe/manifest"

type TasksRenamer interface {
	Apply(original manifest.TaskList) (updated manifest.TaskList)
}

type TasksTimeoutDefaulter interface {
	Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

type TasksEnvVarsDefaulter interface {
	Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

type tasksDefaulter struct {
	runDefaulter                         func(original manifest.Run, defaults Defaults) (updated manifest.Run)
	dockerComposeDefaulter               func(original manifest.DockerCompose, defaults Defaults) (updated manifest.DockerCompose)
	dockerPushDefaulter                  func(original manifest.DockerPush, man manifest.Manifest, defaults Defaults) (updated manifest.DockerPush)
	deployCfDefaulter                    func(original manifest.DeployCF, defaults Defaults, man manifest.Manifest) (updated manifest.DeployCF)
	deployKateeDefaulter                 func(original manifest.DeployKatee, defaults Defaults, man manifest.Manifest) (updated manifest.DeployKatee)
	consumerIntegrationTestTaskDefaulter func(original manifest.ConsumerIntegrationTest, defaults Defaults) (updated manifest.ConsumerIntegrationTest)
	deployMlZipDefaulter                 func(original manifest.DeployMLZip, defaults Defaults) (updated manifest.DeployMLZip)
	deployMlModulesDefaulter             func(original manifest.DeployMLModules, defaults Defaults) (updated manifest.DeployMLModules)
	buildpackDefaulter                   func(original manifest.Buildpack, defaults Defaults) (updated manifest.Buildpack)

	tasksRenamer          TasksRenamer
	tasksTimeoutDefaulter TasksTimeoutDefaulter
	tasksEnvVarsDefaulter TasksEnvVarsDefaulter
}

func NewTaskDefaulter() TasksDefaulter {
	return tasksDefaulter{
		runDefaulter:                         runDefaulter,
		dockerComposeDefaulter:               dockerComposeDefaulter,
		dockerPushDefaulter:                  dockerPushDefaulter,
		deployCfDefaulter:                    deployCfDefaulter,
		deployKateeDefaulter:                 deployKateeDefaulter,
		consumerIntegrationTestTaskDefaulter: consumerIntegration,
		deployMlZipDefaulter:                 deployMlZipDefaulter,
		deployMlModulesDefaulter:             deployMlModuleDefaulter,
		buildpackDefaulter:                   buildpackDefaulter,

		tasksRenamer:          NewTasksRenamer(),
		tasksTimeoutDefaulter: NewTasksTimeoutDefaulter(),
		tasksEnvVarsDefaulter: NewTasksEnvVarsDefaulter(),
	}
}

func (t tasksDefaulter) Apply(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList) {
	tasksWithUniqueName := t.tasksRenamer.Apply(original)

	var tasksWithDefaultsApplied manifest.TaskList
	for _, task := range t.tasksRenamer.Apply(tasksWithUniqueName) {
		var tt manifest.Task
		switch task := task.(type) {
		case manifest.Update:
			tt = task
		case manifest.Run:
			tt = t.runDefaulter(task, defaults)
		case manifest.DockerCompose:
			tt = t.dockerComposeDefaulter(task, defaults)
		case manifest.DockerPush:
			tt = t.dockerPushDefaulter(task, man, defaults)
		case manifest.DockerPushAWS:
			tt = task
		case manifest.DeployCF:
			ppTasks := t.Apply(task.PrePromote, defaults, man)
			task = t.deployCfDefaulter(task, defaults, man)
			task.PrePromote = ppTasks
			tt = task
		case manifest.DeployKatee:
			tt = t.deployKateeDefaulter(task, defaults, man)
		case manifest.Buildpack:
			tt = t.buildpackDefaulter(task, defaults)
		case manifest.ConsumerIntegrationTest:
			tt = t.consumerIntegrationTestTaskDefaulter(task, defaults)
		case manifest.DeployMLModules:
			tt = t.deployMlModulesDefaulter(task, defaults)
		case manifest.DeployMLZip:
			tt = t.deployMlZipDefaulter(task, defaults)
		case manifest.Parallel:
			task.Tasks = t.Apply(task.Tasks, defaults, man)
			tt = task
		case manifest.Sequence:
			task.Tasks = t.Apply(task.Tasks, defaults, man)
			tt = task
		}

		tasksWithDefaultsApplied = append(tasksWithDefaultsApplied, tt)
	}

	tasksWithTimeoutApplied := t.tasksTimeoutDefaulter.Apply(tasksWithDefaultsApplied, defaults)
	tasksWithEnvVarsApplied := t.tasksEnvVarsDefaulter.Apply(tasksWithTimeoutApplied, defaults)

	return tasksWithEnvVarsApplied
}
