package linters

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

type topLevelLinter struct{}

func NewTopLevelLinter() topLevelLinter {
	return topLevelLinter{}
}

func (topLevelLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Halfpipe Manifest"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/"

	if manifest.Team == "" {
		result.AddError(NewErrMissingField("team"))
	} else if strings.ToLower(manifest.Team) != manifest.Team {
		result.AddWarning(NewErrInvalidField("team", "should be lower case"))
	}

	if manifest.Pipeline == "" {
		result.AddError(NewErrMissingField("pipeline"))
	}

	if strings.Contains(manifest.Pipeline, " ") {
		result.AddError(NewErrInvalidField("pipeline", "must not contains spaces!"))
	}

	if (manifest.ArtifactConfig.Bucket != "" && manifest.ArtifactConfig.JSONKey == "") ||
		(manifest.ArtifactConfig.Bucket == "" && manifest.ArtifactConfig.JSONKey != "") {
		result.AddError(NewErrInvalidField("artifact_config", "both 'bucket' and 'json_key' must be specified!"))
	}

	if !(manifest.Platform == "actions" || manifest.Platform == "concourse") {
		result.AddError(NewErrInvalidField("platform", "must be either 'actions' or 'concourse'"))
	}

	return result
}
