package linters

import (
	"regexp"

	"github.com/springernature/halfpipe/manifest"
)

var opsLevelSystemRegex = regexp.MustCompile(`^appl-[0-9]+$`)

type opsLevelLinter struct{}

func NewOpsLevelLinter() opsLevelLinter {
	return opsLevelLinter{}
}

func (opsLevelLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "OpsLevel"
	result.DocsURL = "https://springernature.atlassian.net/wiki/spaces/ENG/pages/600703533/Developer+Portal"

	if manifest.OpsLevel.RelativePath == "" {
		result.Add(ErrOpsLevelNotFound.AsWarning())
		return result
	}

	if manifest.OpsLevel.ParseError != "" {
		result.Add(ErrOpsLevelInvalid.WithValue(manifest.OpsLevel.ParseError).WithValue(manifest.OpsLevel.RelativePath).AsWarning())
		return result
	}

	if !opsLevelSystemRegex.MatchString(manifest.OpsLevel.System) {
		result.Add(NewErrInvalidField("opslevel.system", "must match ^appl-[0-9]+$").WithValue(manifest.OpsLevel.RelativePath).AsWarning())
	}

	return result
}
