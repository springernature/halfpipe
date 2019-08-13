package parallel

import "github.com/springernature/halfpipe/manifest"

type Merger struct {
}

func NewParallelMerger() Merger {
	return Merger{}
}

func (Merger) MergeParallelTasks(tasks manifest.TaskList) (mergedTasks manifest.TaskList) {
	tmpParallel := manifest.Parallel{}
	previousParallelName := ""

	for _, task := range tasks {
		if _, isParallelTask := task.(manifest.Parallel); isParallelTask {
			if len(tmpParallel.Tasks) > 0 {
				mergedTasks = append(mergedTasks, tmpParallel)
			}
			tmpParallel = manifest.Parallel{}
			previousParallelName = ""

			mergedTasks = append(mergedTasks, task)
			continue
		}
		if task.GetParallelGroup().IsSet() {
			currentParallelName := string(task.GetParallelGroup())
			if previousParallelName != currentParallelName {
				if len(tmpParallel.Tasks) > 0 {
					mergedTasks = append(mergedTasks, tmpParallel)
				}
				tmpParallel = manifest.Parallel{}
				previousParallelName = currentParallelName
			}

			tmpParallel.Tasks = append(tmpParallel.Tasks, task)
		} else {
			if len(tmpParallel.Tasks) > 0 {
				mergedTasks = append(mergedTasks, tmpParallel)
			}
			tmpParallel = manifest.Parallel{}
			previousParallelName = ""

			mergedTasks = append(mergedTasks, task)
		}
	}

	if len(tmpParallel.Tasks) > 0 {
		mergedTasks = append(mergedTasks, tmpParallel)
	}

	return
}
