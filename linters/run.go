package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

var ErrScriptMustExistInDockerImage = newError("make sure script is present in the docker image")
var ErrWindowsScriptMustBeExecutable = newError("make sure script is executable")

func LintRunTask(run manifest.Run, fs afero.Afero, os string) (errs []error) {
	if run.Script == "" {
		errs = append(errs, NewErrMissingField("script"))
	} else if strings.HasPrefix(run.Script, `\`) {
		command := strings.Fields(strings.TrimSpace(run.Script))[0]
		commandWithoutSlashPrefix := command[1:]
		errs = append(errs, ErrScriptMustExistInDockerImage.WithFile(commandWithoutSlashPrefix).AsWarning())
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		checkForExecutable := os != "windows"
		if !checkForExecutable {
			errs = append(errs, ErrWindowsScriptMustBeExecutable.WithFile(command).AsWarning())
		}
		if err := CheckFile(fs, command, checkForExecutable); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Retries < 0 || run.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if run.Docker.Image == "" {
		errs = append(errs, NewErrMissingField("docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		errs = append(errs, NewErrMissingField("docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		errs = append(errs, NewErrMissingField("docker.username"))
	}

	return errs
}
