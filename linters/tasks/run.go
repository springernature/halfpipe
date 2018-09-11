package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func LintRunTask(run manifest.Run, fs afero.Afero) (errs []error, warnings []error) {
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		if err := filechecker.CheckFile(fs, command, true); err != nil {
			warnings = append(warnings, err)
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
