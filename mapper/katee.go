package mapper

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

type katee struct {
	fs afero.Afero
}

func (k katee) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	u, err := k.updateTasks(updated.Tasks)
	if err != nil {
		return
	}
	updated.Tasks = u
	return updated, nil
}

func (k katee) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList, err error) {
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
		case manifest.DeployKatee:
			mappedTask, e := k.mapKatee(task)
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

func (k katee) mapKatee(task manifest.DeployKatee) (updated manifest.DeployKatee, err error) {
	updated = task

	content, err := k.fs.ReadFile(task.VelaManifest)
	if err != nil {
		return
	}

	velaManifest, err := manifest.UnmarshalVelaManifest(content)
	if err != nil {
		return
	}

	updated.KateeManifest = velaManifest
	return
}

func NewKateeMapper(fs afero.Afero) Mapper {
	return katee{fs: fs}
}
