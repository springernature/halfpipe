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
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/"

	if manifest.OpsLevel.System != "" && !opsLevelSystemRegex.MatchString(manifest.OpsLevel.System) {
		result.Add(NewErrInvalidField("opslevel.system", "must match ^appl-[0-9]+$").AsWarning())
	}

	return result
}
