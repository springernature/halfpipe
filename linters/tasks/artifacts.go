package tasks

import (
	"fmt"
	"github.com/pkg/errors"
	errors2 "github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"regexp"
)

func LintArtifacts(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
	if currentTask.ReadsFromArtifacts() && !previousTasksSavesArtifact(previousTasks) {
		errs = append(errs, errors.New("reads from saved artifacts, but there are no previous tasks that saves any"))
	}

	var saveArtifacts []string
	var deployArtifact string
	switch currentTask := currentTask.(type) {
	case manifest.Run:
		saveArtifacts = currentTask.SaveArtifacts
	case manifest.DockerCompose:
		saveArtifacts = currentTask.SaveArtifacts
	case manifest.DeployCF:
		deployArtifact = currentTask.DeployArtifact
	}

	environmentVariableNameRegex := regexp.MustCompile(`\$[a-zA-Z0-9_]*`)

	for _, saveArtifact := range saveArtifacts {
		if environmentVariableNameRegex.Match([]byte(saveArtifact)) {
			errs = append(errs, errors2.NewInvalidField("save_artifact", fmt.Sprintf("you are not allowed to refer to environment variables: '%s'", saveArtifact)))
		}
	}

	if deployArtifact != "" {
		if environmentVariableNameRegex.Match([]byte(deployArtifact)) {
			errs = append(errs, errors2.NewInvalidField("deploy_artifact", fmt.Sprintf("you are not allowed to refer to environment variables: '%s'", deployArtifact)))
		}
	}

	return errs, warnings
}

func previousTasksSavesArtifact(tasks []manifest.Task) bool {
	for _, task := range tasks {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}
