package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

var WarnScriptMustExistInDockerImage = func(script string) error {
	return fmt.Errorf("make sure '%s' is availible in the docker image you have specified", script)
}

func LintRunTask(run manifest.Run, fs afero.Afero) (errs []error, warnings []error) {
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	} else if strings.HasPrefix(run.Script, `\`) {
		command := strings.Fields(strings.TrimSpace(run.Script))[0]
		commandWithoutSlashPrefix := command[1:]
		warnings = append(warnings, WarnScriptMustExistInDockerImage(commandWithoutSlashPrefix))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		if err := filechecker.CheckFile(fs, command, true); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Retries < 0 || run.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if run.Docker.Image == "" {
		errs = append(errs, errors.NewMissingField("docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		errs = append(errs, errors.NewMissingField("docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		errs = append(errs, errors.NewMissingField("docker.username"))
	}

	return
}
