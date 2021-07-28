package concourse

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) automaticCDCConfig(task manifest.AutomaticCDC) atc.JobConfig {
	jobConfig := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	return jobConfig
}
