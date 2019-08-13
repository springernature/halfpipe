package parallel

import "github.com/springernature/halfpipe/manifest"

type Merger struct {
}

func NewParallelMerger() Merger {
	return Merger{}
}

func (m Merger) removeParallelGroup(task manifest.Task) (fixed manifest.Task) {
	switch task := task.(type) {
	case manifest.ConsumerIntegrationTest:
		task.Parallel = ""
		fixed = task
	case manifest.DeployCF:
		task.Parallel = ""
		fixed = task
	case manifest.DeployMLModules:
		task.Parallel = ""
		fixed = task
	case manifest.DeployMLZip:
		task.Parallel = ""
		fixed = task
	case manifest.DockerCompose:
		task.Parallel = ""
		fixed = task
	case manifest.DockerPush:
		task.Parallel = ""
		fixed = task
	case manifest.Run:
		task.Parallel = ""
		fixed = task
	case manifest.Update, manifest.Parallel:
		fixed = task
	}

	return
}

func (m Merger) MergeParallelTasks(tasks manifest.TaskList) (mergedTasks manifest.TaskList) {
	tmpParallel := manifest.Parallel{}
	previousParallelName := ""

	for _, task := range tasks {
		if _, isParallelTask := task.(manifest.Parallel); isParallelTask {
			if len(tmpParallel.Tasks) > 0 {
				mergedTasks = append(mergedTasks, tmpParallel)
			}
			tmpParallel = manifest.Parallel{}
			previousParallelName = ""

			mergedTasks = append(mergedTasks, m.removeParallelGroup(task))
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

			tmpParallel.Tasks = append(tmpParallel.Tasks, m.removeParallelGroup(task))
		} else {
			if len(tmpParallel.Tasks) > 0 {
				mergedTasks = append(mergedTasks, tmpParallel)
			}
			tmpParallel = manifest.Parallel{}
			previousParallelName = ""

			mergedTasks = append(mergedTasks, m.removeParallelGroup(task))
		}
	}

	if len(tmpParallel.Tasks) > 0 {
		mergedTasks = append(mergedTasks, tmpParallel)
	}

	return
}
