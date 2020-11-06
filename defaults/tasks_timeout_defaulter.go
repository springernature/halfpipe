package defaults

import "github.com/springernature/halfpipe/manifest"

type tasksTimeoutDefaulter struct {
}

func NewTasksTimeoutDefaulter() TasksTimeoutDefaulter {
	return tasksTimeoutDefaulter{}
}

func (t tasksTimeoutDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
	for _, task := range original {
		var tt manifest.Task
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task
		case manifest.Sequence:
			task.Tasks = t.Apply(task.Tasks, defaults)
			tt = task
		default:
			tt = task
			if task.GetTimeout() == "" {
				tt = tt.SetTimeout(defaults.Aux.Timeout)
			}

			if deployTask, isDeployTask := tt.(manifest.DeployCF); isDeployTask {
				deployTask.PrePromote = t.Apply(deployTask.PrePromote, defaults)
				tt = deployTask
			}
		}
		updated = append(updated, tt)
	}

	return updated
}
