package defaults

import (
	"github.com/springernature/halfpipe/manifest"
)

type buildHistoryDefaulter struct {
}

func NewBuildHistoryDefaulter() BuildHistoryDefaulter {
	return buildHistoryDefaulter{}
}

func (t buildHistoryDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {

	//tasksWithUniqueName := t.tasksRenamer.Apply(original)
	//
	//var tasksWithDefaultsApplied manifest.TaskList
	for _, task := range original {
		switch task := task.(type) {
		//parallel
		//sequence
		case manifest.Run,
			manifest.DockerCompose,
			manifest.DeployCF,
			manifest.DockerPush,
			manifest.ConsumerIntegrationTest,
			manifest.DeployMLZip,
			manifest.DeployMLModules,
			manifest.Update:
			if task.GetBuildHistory() == 0 {
				updated = append(updated, task.SetBuildHistory(defaults.BuildHistory))
			} else {
				updated = append(updated, task)
			}
		case manifest.Parallel:
			task.Tasks = t.Apply(task.Tasks, defaults)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = t.Apply(task.Tasks, defaults)
			updated = append(updated, task)
		}
	}

	return
}
