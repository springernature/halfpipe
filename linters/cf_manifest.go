package linters

import (
	"fmt"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type cfManifestLinter struct {
	Fs afero.Afero
}

func NewCfManifestLinter(fs afero.Afero) cfManifestLinter {
	return cfManifestLinter{fs}
}

func (linter cfManifestLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "CF Manifest Linter"
	var manifestPaths []string

	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployCF:
			manifestPaths = append(manifestPaths, t.Manifest)
		}
	}

	for _, manifestPath := range manifestPaths {
		apps, err := linter.readManifest(manifestPath)

		if err != nil {
			result.AddError(err)
			return
		}

		if len(apps) != 1 {
			result.AddError(errors.NewTooManyAppsError(manifestPath, "Manifest must have 1 application defined"))
		}
	}
	return
}

//partly stolen from code.cloudfoundry.org/cli/util/manifest
func (linter cfManifestLinter) readManifest(pathToManifest string) ([]cfManifest.Application, error) {
	raw, err := linter.Fs.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	var man cfManifest.Manifest
	err = yaml.Unmarshal(raw, &man)
	if err != nil {
		errInvalidManifest := fmt.Errorf("error parsing Cf Manifest, manifest %q is invalid", pathToManifest)
		return nil, errInvalidManifest
	}

	return man.Applications, err
}
