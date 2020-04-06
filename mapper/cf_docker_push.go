package mapper

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

type cFDockerPush struct {
}

func (c cFDockerPush) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	u, err := c.updateTasks(updated.Tasks)
	if err != nil {
		return
	}
	updated.Tasks = u
	return updated, nil
}

func (c cFDockerPush) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList, err error) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			u, e := c.updateTasks(task.Tasks)
			if e != nil {
				err = e
				return
			}
			task.Tasks = u
			updated = append(updated, task)
		case manifest.Sequence:
			u, e := c.updateTasks(task.Tasks)
			if e != nil {
				err = e
				return
			}
			task.Tasks = u
			updated = append(updated, task)
		case manifest.DeployCF:
			cf, e := c.setIsDockerPush(task)
			if e != nil {
				err = e
				return
			}
			updated = append(updated, cf)
		default:
			updated = append(updated, task)
		}
	}
	return updated, err
}

func (c cFDockerPush) setIsDockerPush(cf manifest.DeployCF) (updated manifest.DeployCF, err error) {
	updated = cf
	if strings.HasPrefix(cf.Manifest, "../") {
		return
	}

	apps, err := cfManifest.ReadAndInterpolateManifest(cf.Manifest, nil, nil)
	if err != nil {
		return
	}

	if apps[0].DockerImage != "" {
		updated.IsDockerPush = true
	}
	return
}

func NewCFDockerPushMapper() Mapper {
	return cFDockerPush{}
}
