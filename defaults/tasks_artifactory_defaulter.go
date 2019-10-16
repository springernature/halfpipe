package defaults

import "github.com/springernature/halfpipe/manifest"

type tasksArtifactoryVarsDefaulter struct {
}

func NewTasksArtifactoryVarsDefaulter() TasksArtifactoryVarsDefaulter {
	return tasksArtifactoryVarsDefaulter{}
}

func (t tasksArtifactoryVarsDefaulter) addDefaultsToVars(vars manifest.Vars, defaults Defaults) manifest.Vars {
	return vars.
		SetVar("ARTIFACTORY_URL", defaults.ArtifactoryURL).
		SetVar("ARTIFACTORY_USERNAME", defaults.ArtifactoryUsername).
		SetVar("ARTIFACTORY_PASSWORD", defaults.ArtifactoryPassword)
}

func (t tasksArtifactoryVarsDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
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

	return
}
