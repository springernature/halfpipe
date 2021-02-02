package mapper

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

type cf struct {
}

func (c cf) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	u, err := c.updateTasks(updated.Tasks)
	if err != nil {
		return
	}
	updated.Tasks = u
	return updated, nil
}

func (c cf) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList, err error) {
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
			cf, e := c.mapCf(task)
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

func (c cf) mapCf(cf manifest.DeployCF) (updated manifest.DeployCF, err error) {
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

	updated.CfApplication = apps[0]
	return
}

func NewCfMapper() Mapper {
	return cf{}
}
