package linters

import (
	"strings"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

type topLevelLinter struct{}

func NewTopLevelLinter() topLevelLinter {
	return topLevelLinter{}
}

func (topLevelLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Halfpipe Manifest"
	result.DocsURL = "https://docs.halfpipe.io/manifest/"

	if manifest.Team == "" {
		result.AddError(linterrors.NewMissingField("team"))
	} else if strings.ToLower(manifest.Team) != manifest.Team {
		result.AddWarning(linterrors.NewInvalidField("team", "team should be lower case"))
	}

	if manifest.Pipeline == "" {
		result.AddError(linterrors.NewMissingField("pipeline"))
	}

	if strings.Contains(manifest.Pipeline, " ") {
		result.AddError(linterrors.NewInvalidField("pipeline", "pipeline name must not contains spaces!"))
	}

	if (manifest.ArtifactConfig.Bucket != "" && manifest.ArtifactConfig.JSONKey == "") ||
		(manifest.ArtifactConfig.Bucket == "" && manifest.ArtifactConfig.JSONKey != "") {
		result.AddError(linterrors.NewInvalidField("artifact_config", "both 'bucket' and 'json_key' must be specified!"))
	}

	return result
}
