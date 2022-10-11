package migrate

import (
	"errors"
	"fmt"
	"github.com/springernature/halfpipe/linters"
	"reflect"

	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
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

var ErrLintingOriginalManifest = errors.New("linting original manifest failed")
var ErrLintingMigratedManifest = errors.New("linting migrated manifest failed")

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

func (m migrator) Migrate(man manifest.Manifest) (migratedManifest manifest.Manifest, migratedYaml []byte, lintResults linters.LintResults, migrated bool, err error) {
	// Checking that the original manifest is ok
	response := m.controller.Process(man)
	lintResults = response.LintResults
	if response.LintResults.HasErrors() {
		err = ErrLintingOriginalManifest
		return
	}

	// migrating the manifest
	tmpMigratedManifest := man

	// if migrated and original manifest is the same
	if reflect.DeepEqual(tmpMigratedManifest, man) {
		migratedManifest = tmpMigratedManifest
		migrated = false
		return
	}

	// check the migrated manifest
	responseMigrated := m.controller.Process(tmpMigratedManifest)
	if responseMigrated.LintResults.HasErrors() {
		migratedManifest = man
		err = ErrLintingMigratedManifest
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
