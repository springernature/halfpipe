package mapper

import (
	"fmt"

	"github.com/springernature/halfpipe/manifest"
)

type uploadSlos struct{}

func (k uploadSlos) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	u, err := k.updateTasks(updated.Tasks)
	if err != nil {
		return
	}
	updated.Tasks = u
	return updated, nil
}

func (k uploadSlos) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList, err error) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			u, e := k.updateTasks(task.Tasks)
			if e != nil {
				err = e
				return
			}
			task.Tasks = u
			updated = append(updated, task)
		case manifest.Sequence:
			u, e := k.updateTasks(task.Tasks)
			if e != nil {
				err = e
				return
			}
			task.Tasks = u
			updated = append(updated, task)
		case manifest.UploadSLOs:
			mappedTask, e := k.mapUploadSLOs(task)
			if e != nil {
				err = e
				return
			}
			updated = append(updated, mappedTask)
		default:
			updated = append(updated, task)
		}
	}
	return updated, err
}

func (k uploadSlos) mapUploadSLOs(task manifest.UploadSLOs) (mapped manifest.Run, err error) {
	mapped.Docker.Image = "eu.gcr.io/halfpipe-io/engineering-enablement/o11ytool:0.2.10"
	mapped.Script = fmt.Sprintf(`\upload-slos -i %s`, task.Folder)
	mapped.Name = task.Name
	mapped.Vars = task.GetVars()
	return mapped, nil
}

func NewUploadSlosMapper() Mapper {
	return uploadSlos{}
}
