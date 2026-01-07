package linters

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/springernature/halfpipe/manifest"
)

func LintArtifacts(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error) {
	if currentTask.ReadsFromArtifacts() && !previousTasksSavesArtifact(previousTasks) {
		errs = append(errs, ErrReadsFromSavedArtifacts)
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
			errs = append(errs, NewErrInvalidField("save_artifact", fmt.Sprintf("you are not allowed to refer to environment variables: '%s'", saveArtifact)))
		}

		if saveArtifact == "" {
			errs = append(errs, NewErrInvalidField("save_artifact", "empty path"))
		}
	}

	if deployArtifact != "" {
		if environmentVariableNameRegex.Match([]byte(deployArtifact)) {
			errs = append(errs, NewErrInvalidField("deploy_artifact", fmt.Sprintf("you are not allowed to refer to environment variables: '%s'", deployArtifact)))
		}
	}

	return errs
}

func previousTasksSavesArtifact(tasks []manifest.Task) bool {
	return slices.ContainsFunc(tasks, func(t manifest.Task) bool { return t.SavesArtifacts() })
}
