package defaults

import "github.com/springernature/halfpipe/manifest"

type tasksEnvVarsDefaulter struct {
}

func NewTasksEnvVarsDefaulter() TasksEnvVarsDefaulter {
	return tasksEnvVarsDefaulter{}
}

func (t tasksEnvVarsDefaulter) addDefaultsToVars(vars manifest.Vars, defaults Defaults) manifest.Vars {
	if defaults.Artifactory == (ArtifactoryDefaults{}) {
		return vars
	}
	if vars == nil {
		vars = make(manifest.Vars)
	}
	vars["ARTIFACTORY_URL"] = defaults.Artifactory.URL
	vars["ARTIFACTORY_USERNAME"] = defaults.Artifactory.Username
	vars["ARTIFACTORY_PASSWORD"] = defaults.Artifactory.Password
	vars["RUNNING_IN_CI"] = "true"
	vars["CI"] = "true"

	return vars
}

func (t tasksEnvVarsDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
	for _, task := range original {
		var tt manifest.Task
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task
		case manifest.Sequence:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task

		case manifest.DockerCompose:
			task.Vars = t.addDefaultsToVars(task.Vars, defaults)
			tt = task
		case manifest.Run:
			task.Vars = t.addDefaultsToVars(task.Vars, defaults)
			tt = task
		case manifest.DockerPush:
			task.Vars = t.addDefaultsToVars(task.Vars, defaults)
			tt = task
		case manifest.DeployCF:
			task.PrePromote = t.Apply(task.PrePromote, defaults)
			tt = task
		case manifest.ConsumerIntegrationTest:
			task.Vars = t.addDefaultsToVars(task.Vars, defaults)
			tt = task
		default:
			tt = task
		}
		updated = append(updated, tt)
	}

	return updated
}
