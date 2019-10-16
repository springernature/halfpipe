package defaults

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

type tasksRenamer struct {
}

func NewTasksRenamer() TasksRenamer {
	return tasksRenamer{}
}

func getUniqueName(name string, previousNames []string, counter int) string {
	candidate := name
	if counter > 0 {
		candidate = fmt.Sprintf("%s (%v)", name, counter)
	}

	for _, previousName := range previousNames {
		if previousName == candidate {
			return getUniqueName(name, previousNames, counter+1)
		}
	}

	return candidate

}

func uniqueName(name string, previousNames []string) string {
	return getUniqueName(name, previousNames, 0)
}

func uniqueifyNames(tasks manifest.TaskList) manifest.TaskList {
	var previousNames []string
	var taskSwitcher func(tasks manifest.TaskList) manifest.TaskList
	taskSwitcher = func(tasks manifest.TaskList) manifest.TaskList {
		var updatedTasks manifest.TaskList
		for _, task := range tasks {
			switch task := task.(type) {
			case manifest.Parallel:
				task.Tasks = taskSwitcher(task.Tasks)
				updatedTasks = append(updatedTasks, task)
			case manifest.Sequence:
				task.Tasks = taskSwitcher(task.Tasks)
				updatedTasks = append(updatedTasks, task)
			default:
				newName := uniqueName(task.GetName(), previousNames)
				previousNames = append(previousNames, newName)
				if deployCf, ok := task.(manifest.DeployCF); ok {
					deployCf.PrePromote = uniqueifyNames(deployCf.PrePromote)
					task = deployCf
				}
				updatedTasks = append(updatedTasks, task.SetName(newName))
			}
		}
		return updatedTasks
	}
	return taskSwitcher(tasks)
}

func (t tasksRenamer) Apply(original manifest.TaskList) (updated manifest.TaskList) {
	return uniqueifyNames(original)
}
