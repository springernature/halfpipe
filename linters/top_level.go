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
		result.Add(NewErrMissingField("team"))
	} else if strings.ToLower(manifest.Team) != manifest.Team {
		result.Add(NewErrInvalidField("team", "should be lower case").AsWarning())
	}

	if manifest.Pipeline == "" {
		result.Add(NewErrMissingField("pipeline"))
	}

	if strings.Contains(manifest.Pipeline, " ") {
		result.Add(NewErrInvalidField("pipeline", "must not contains spaces!"))
	}

	if (manifest.ArtifactConfig.Bucket != "" && manifest.ArtifactConfig.JSONKey == "") ||
		(manifest.ArtifactConfig.Bucket == "" && manifest.ArtifactConfig.JSONKey != "") {
		result.Add(NewErrInvalidField("artifact_config", "both 'bucket' and 'json_key' must be specified!"))
	}

	if !(manifest.Platform == "actions" || manifest.Platform == "concourse") {
		result.Add(NewErrInvalidField("platform", "must be either 'actions' or 'concourse'"))
	}

	if manifest.SlackSuccessMessage != "" {
		result.Add(ErrSlackSuccessMessageFieldDeprecated.AsWarning())
	}

	if manifest.SlackFailureMessage != "" {
		result.Add(ErrSlackFailureMessageFieldDeprecated.AsWarning())
	}

	nots := append(manifest.Notifications.Failure, manifest.Notifications.Success...)
	for _, n := range nots {
		if n.Slack != "" && n.Teams != "" {
			result.Add(ErrOnlySlackOrTeamsAllowed)
		}
	}

	return result
}
