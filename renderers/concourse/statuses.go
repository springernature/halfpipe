package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func statusesOnFailurePlan(man manifest.Manifest, task manifest.Task) atc.Step {
	return statusesPlan("failure", man, task)
}

func statusesOnSuccessPlan(man manifest.Manifest, task manifest.Task) atc.Step {
	return statusesPlan("success", man, task)
}

func statusesPlan(status string, man manifest.Manifest, task manifest.Task) atc.Step {

	return atc.Step{
		Config: &atc.PutStep{
			Name: githubStatusesResourceName,
			Params: atc.Params{
				"state":   status,
				"context": fmt.Sprintf("%s/%s", man.PipelineName(), task.GetName()),
				"path":    gitDir,
			},
		},
	}
}
