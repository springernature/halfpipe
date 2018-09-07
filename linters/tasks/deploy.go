package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"strings"
	"time"
)

func LintDeployCFTask(cf manifest.DeployCF, taskID string, fs afero.Afero) (errs []error, warnings []error) {
	if cf.API == "" {
		errs = append(errs, errors.NewMissingField(taskID+" deploy-cf.api"))
	}
	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField(taskID+" deploy-cf.space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField(taskID+" deploy-cf.org"))
	}
	if cf.TestDomain == "" {
		_, found := defaults.DefaultValues.CfTestDomains[cf.API]
		if cf.API != "" && !found {
			errs = append(errs, errors.NewMissingField(taskID+" deploy-cf.testDomain"))
		}
	}

	if cf.Timeout != "" {
		_, err := time.ParseDuration(cf.Timeout)
		if err != nil {
			errs = append(errs, errors.NewInvalidField(taskID+" deploy-cf.timeout", err.Error()))
		}
	}

	if cf.Retries < 0 || cf.Retries > 5 {
		errs = append(errs, errors.NewInvalidField(taskID+" deploy-cf.retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(cf.Manifest, "../artifacts/") {
		warnings = append(warnings, errors.NewFileError(cf.Manifest, "this file must be saved as an artifact in a previous task"))
	} else if err := filechecker.CheckFile(fs, cf.Manifest, false); err != nil {
		errs = append(errs, err)
	}

	for i, prePromote := range cf.PrePromote {
		ppTaskID := fmt.Sprintf("%s.pre_promote[%v]", taskID, i)
		switch task := prePromote.(type) {
		case manifest.Run:
			if task.ManualTrigger == true {
				errs = append(errs, errors.NewInvalidField(ppTaskID+" run.manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
			}
			if task.Parallel {
				errs = append(errs, errors.NewInvalidField(ppTaskID+" run.passed", "You are not allowed to set 'passed' inside a pre promote task"))
			}
		case manifest.DockerCompose:
			if task.ManualTrigger == true {
				errs = append(errs, errors.NewInvalidField(ppTaskID+" docker-compose.manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
			}
			if task.Parallel {
				errs = append(errs, errors.NewInvalidField(ppTaskID+" docker-compose.passed", "You are not allowed to set 'passed' inside a pre promote task"))
			}
		case manifest.DockerPush, manifest.DeployCF:
			errs = append(errs, errors.NewInvalidField(ppTaskID+" run.type", "You are not allowed to have a 'deploy-cf' or 'docker-push' task as a pre promote"))
		}

	}

	return
}
