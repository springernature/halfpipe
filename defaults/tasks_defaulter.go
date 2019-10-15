package defaults

import "github.com/springernature/halfpipe/manifest"

type tasksDefaulter struct {
	runDefaulter                         func(original manifest.Run, defaults DefaultsNew) (updated manifest.Run)
	dockerComposeDefaulter               func(original manifest.DockerCompose, defaults DefaultsNew) (updated manifest.DockerCompose)
	dockerPushDefaulter                  func(original manifest.DockerPush, defaults DefaultsNew) (updated manifest.DockerPush)
	deployCfDefaulter                    func(original manifest.DeployCF, defaults DefaultsNew) (updated manifest.DeployCF)
	consumerIntegrationTestTaskDefaulter func(original manifest.ConsumerIntegrationTest, defaults DefaultsNew) (updated manifest.ConsumerIntegrationTest)
	deployMlZipDefaulter                 func(original manifest.DeployMLZip, defaults DefaultsNew) (updated manifest.DeployMLZip)
	deployMlModulesDefaulter             func(original manifest.DeployMLModules, defaults DefaultsNew) (updated manifest.DeployMLModules)
}

func NewTaskDefaulter() TasksDefaulter {
	return tasksDefaulter{

	}
}

func (t tasksDefaulter) Apply(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList) {
	for _, task := range original {
		var tt manifest.Task
		switch task := task.(type) {
		case manifest.Update:
			tt = task
		case manifest.Run:
			tt = t.runDefaulter(task, defaults)
		case manifest.DockerCompose:
			tt = t.dockerComposeDefaulter(task, defaults)
		case manifest.DockerPush:
			tt = t.dockerPushDefaulter(task, defaults)
		case manifest.DeployCF:
			ppTasks := t.Apply(task.PrePromote, defaults)
			task = t.deployCfDefaulter(task, defaults)
			task.PrePromote = ppTasks
			tt = task
		case manifest.ConsumerIntegrationTest:
			tt = t.consumerIntegrationTestTaskDefaulter(task, defaults)
		case manifest.DeployMLModules:
			tt = t.deployMlModulesDefaulter(task, defaults)
		case manifest.DeployMLZip:
			tt = t.deployMlZipDefaulter(task, defaults)
		case manifest.Parallel:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task
		case manifest.Sequence:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task
		}
		updated = append(updated, tt)
	}

	return
}
