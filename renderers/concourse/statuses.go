package concourse

import (
	"github.com/concourse/concourse/atc"
)

func statusesOnFailurePlan() atc.Step {
	return statusesPlan("failure")
}

func statusesOnSuccessPlan() atc.Step {
	return statusesPlan("success")
}

func statusesPlan(status string) atc.Step {

	return atc.Step{
		Config: &atc.PutStep{
			Name: githubStatusesResourceName,
			Params: atc.Params{
				"state": status,
			},
			NoGet: true,
		},
	}
}
