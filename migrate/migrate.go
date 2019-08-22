package migrate

import (
	"errors"
	"fmt"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/parallel"
	"github.com/springernature/halfpipe/triggers"
	"reflect"
)

type ParseFunc func(manifestYaml string) (manifest.Manifest, []error)
type RenderFunc func(manifest manifest.Manifest) (y []byte, err error)

type Migrator interface {
	Migrate(man manifest.Manifest) manifest.Manifest
}

func NewMigrator(controller halfpipe.Controller, parseFunc ParseFunc, renderFunc RenderFunc) migrator {
	return migrator{
		controller: controller,
		parseFunc:  parseFunc,
		renderFunc: renderFunc,
	}
}

type migrator struct {
	controller halfpipe.Controller
	parseFunc  ParseFunc
	renderFunc RenderFunc
}

var LintingOriginalManifestErr = errors.New("linting original manifest failed")
var LintingMigratedManifestErr = errors.New("linting migrated manifest failed")

var FailedToRenderMigratedManifestToYamlErr = func(err error, man manifest.Manifest) error {
	return fmt.Errorf("%s: \n%+v", err, man)
}

var FailedToParseMigratedManifestYamlErr = func(errs []error, yM string) error {
	errorStr := "Failed to parse manifest:"
	for _, err := range errs {
		errorStr = fmt.Sprintf("%s\n%s", errorStr, err.Error())
	}

	return fmt.Errorf("%s\n\n%s", errorStr, yM)
}

var ParsedMigratedManifestAndMigratedManifestIsNotTheSameErr = func(migratedManifest, parsedFromMigratedYaml manifest.Manifest) error {
	errStr := fmt.Sprintf(`parsed manifest from migrated manifest yaml is not the same! This should never happen!
migratedManifest:
%+v
parsedMigratedManifestYaml:
%+v
`, migratedManifest, parsedFromMigratedYaml)
	return errors.New(errStr)
}

func (m migrator) Migrate(man manifest.Manifest) (migratedManifest manifest.Manifest, migratedYaml []byte, lintResults result.LintResults, err error, migrated bool) {
	// Checking that the original manifest is ok
	_, lintResults = m.controller.Process(man)
	if lintResults.HasErrors() {
		err = LintingOriginalManifestErr
		return
	}

	// migrating the manifest
	tmpMigratedManifest := man
	tmpMigratedManifest = triggers.NewTriggersTranslator().Translate(tmpMigratedManifest)
	tmpMigratedManifest.Tasks = parallel.NewParallelMerger().MergeParallelTasks(tmpMigratedManifest.Tasks)

	// if migrated and original manifest is the same
	if reflect.DeepEqual(tmpMigratedManifest, man) {
		migratedManifest = tmpMigratedManifest
		migrated = false
		return
	}

	// check the migrated manifest
	_, lintResults = m.controller.Process(tmpMigratedManifest)
	if lintResults.HasErrors() {
		migratedManifest = man
		err = LintingMigratedManifestErr
		return
	}

	migratedManifestYaml, err := m.renderFunc(tmpMigratedManifest)
	if err != nil {
		err = FailedToRenderMigratedManifestToYamlErr(err, tmpMigratedManifest)
		return
	}

	parsedMigratedManifest, errs := m.parseFunc(string(migratedManifestYaml))
	if len(errs) > 0 {
		err = FailedToParseMigratedManifestYamlErr(errs, string(migratedManifestYaml))
		return
	}

	if !reflect.DeepEqual(tmpMigratedManifest, parsedMigratedManifest) {
		err = ParsedMigratedManifestAndMigratedManifestIsNotTheSameErr(tmpMigratedManifest, parsedMigratedManifest)
		return
	}

	migratedManifest = tmpMigratedManifest
	migratedYaml = migratedManifestYaml
	migrated = true
	return
}
