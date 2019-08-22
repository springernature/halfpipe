package cmds

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/parallel"
	"github.com/springernature/halfpipe/triggers"
	"github.com/tcnksm/go-gitconfig"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/sync"
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
		defaults.NewDefaulter(projectData),
		parallel.NewParallelMerger(),
		triggers.NewTriggersTranslator(),
		[]linters.Linter{
			linters.NewTopLevelLinter(),
			linters.NewTriggersLinter(fs, currentDir, project.BranchResolver, gitconfig.OriginURL),
			linters.NewSecretsLinter(manifest.NewSecretValidator()),
			linters.NewTasksLinter(fs, runtime.GOOS),
			linters.NewCfManifestLinter(cfManifest.ReadAndInterpolateManifest),
			linters.NewFeatureToggleLinter(manifest.AvailableFeatureToggles),
		},
		pipeline.NewPipeline(cfManifest.ReadAndInterpolateManifest, fs),
	)

}

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render concourse pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
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
			printErrAndResultAndExitOnError(nil, result.LintResults{result.NewLintResult("Halfpipe", "https://docs.halfpipe.io/manifest/", manErrors, nil)})
		}

		pipelineConfig, lintResults := createController(projectData, fs, currentDir).Process(man)
		printErrAndResultAndExitOnError(nil, lintResults)

		pipelineString, renderError := pipeline.ToString(pipelineConfig)
		if renderError != nil {
			printErr(fmt.Errorf("%s\n%s", err, renderError))
			os.Exit(1)
		}

		fmt.Println(pipelineString)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func printErr(err error) {
	fmt.Fprintln(os.Stderr, err) // nolint: gas
}
