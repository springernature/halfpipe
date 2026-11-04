package mapper

import (
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

type observabilityProvisioning struct{}

func (k observabilityProvisioning) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	u, err := k.updateTasks(updated.Tasks)
	if err != nil {
		return
	}
	updated.Tasks = u
	return updated, nil
}

func (k observabilityProvisioning) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList, err error) {
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
		case manifest.ObservabilityProvisioning:
			mappedTask, e := k.mapObservabilityProvisioning(task)
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

func (k observabilityProvisioning) mapObservabilityProvisioning(task manifest.ObservabilityProvisioning) (mapped manifest.Run, err error) {
	mapped.Name = task.Name
	mapped.Vars = task.GetVars()
	mapped.Docker = manifest.Docker{
		Image:    "eu.gcr.io/halfpipe-io/engineering-enablement/o11ytool:latest",
		Username: "oauth2accesstoken",
		Password: config.VaultSecrets.GARToken,
	}

	args := []string{`\provisioning`}
	addOptionalArg := func(flag string, value string) {
		if value != "" {
			args = append(args, flag, value)
		}
	}
	addOptionalArg("--slos", task.SLOs)
	addOptionalArg("--metric-rules", task.MetricRules)
	addOptionalArg("--alertmanager", task.AlertManager)
	addOptionalArg("--dashboards", task.Dashboards)
	mapped.Script = strings.Join(args, " ")

	return mapped, nil
}

func NewObservabilityProvisioningMapper() Mapper {
	return observabilityProvisioning{}
}
