package linters

import (
	"fmt"
	"regexp"

	"github.com/springernature/halfpipe/manifest"
)

var opsLevelSystemRegex = regexp.MustCompile(`^APPL-[0-9]+$`)

type opsLevelLinter struct{}

func NewOpsLevelLinter() opsLevelLinter {
	return opsLevelLinter{}
}

func (opsLevelLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "OpsLevel"
	result.DocsURL = "https://springernature.atlassian.net/wiki/spaces/ENG/pages/600703533/Developer+Portal"

	if manifest.OpsLevel.RelativePath == "" {
		result.Add(ErrOpsLevelNotFound)
		return result
	}

	if manifest.OpsLevel.ParseError != "" {
		result.Add(ErrOpsLevelInvalid.WithValue(manifest.OpsLevel.ParseError).WithValue(manifest.OpsLevel.RelativePath))
		return result
	}

	if !opsLevelSystemRegex.MatchString(manifest.OpsLevel.System) {
		result.Add(NewErrInvalidField(
			"component.system",
			fmt.Sprintf("must match %s", opsLevelSystemRegex)).WithValue(manifest.OpsLevel.RelativePath))
	}

	return result
}
