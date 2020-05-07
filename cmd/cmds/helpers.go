package cmds

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/sync"
	"github.com/tcnksm/go-gitconfig"
	"os"
	"path"
	"runtime"
)

var checkVersion = func() (err error) {
	currentVersion, err := config.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion, sync.ResolveLatestVersionFromArtifactory)
	err = syncer.Check()
	return
}

func getManifest(fs afero.Afero, currentDir, halfpipeFilePath string) (man manifest.Manifest, errors []error) {
	yaml, err := filechecker.ReadFile(fs, path.Join(currentDir, halfpipeFilePath))
	if err != nil {
		errors = append(errors, err)
		return
	}

	man, errs := manifest.Parse(yaml)
	if len(errs) != 0 {
		errors = append(errors, errs...)
		return
	}

	return
}

func printErr(err error) {
	fmt.Fprintln(os.Stderr, err) // nolint: gas
}

func printErrAndResultAndExitOnError(err error, lintResults result.LintResults) {
	if lintResults.HasWarnings() && !lintResults.HasErrors() && err == nil {
		printErr(fmt.Errorf(lintResults.Error()))
		return
	}

	if err != nil || lintResults.HasErrors() {
		if err != nil {
			printErr(err)
		}

		if lintResults.HasErrors() {
			printErr(fmt.Errorf(lintResults.Error()))
		}

		os.Exit(1)
	}
}

func createController(projectData project.Data, fs afero.Afero, currentDir string) halfpipe.Controller {
	return halfpipe.NewController(
		defaults.New(projectData),
		mapper.New(),
		[]linters.Linter{
			linters.NewTopLevelLinter(),
			linters.NewTriggersLinter(fs, currentDir, project.BranchResolver, gitconfig.OriginURL, config.DeprecatedDockerRegistries),
			linters.NewSecretsLinter(manifest.NewSecretValidator()),
			linters.NewTasksLinter(fs, runtime.GOOS, config.DeprecatedCFApis),
			linters.NewCfManifestLinter(cfManifest.ReadAndInterpolateManifest),
			linters.NewFeatureToggleLinter(manifest.AvailableFeatureToggles),
			linters.NewDeprecatedDockerRegistriesLinter(fs, config.DeprecatedDockerRegistries),
		},
		pipeline.NewPipeline(cfManifest.ReadAndInterpolateManifest, fs),
	)

}

func getManifestAndCreateController() (manifest.Manifest, halfpipe.Controller) {
	if err := checkVersion(); err != nil {
		printErr(err)
		os.Exit(1)
	}

	fs := afero.Afero{Fs: afero.NewOsFs()}

	currentDir, err := os.Getwd()
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	man, manErrors := getManifest(fs, currentDir, projectData.HalfpipeFilePath)
	if len(manErrors) > 0 {
		printErrAndResultAndExitOnError(nil, result.LintResults{result.NewLintResult("Halfpipe Manifest", "https://docs.halfpipe.io/manifest/", manErrors, nil)})
	}

	controller := createController(projectData, fs, currentDir)

	return man, controller
}
