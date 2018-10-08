package cmds

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	man "github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/sync"
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

		projectData, err := project.NewProjectResolver(fs).Parse(currentDir)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		ctrl := halfpipe.Controller{
			Fs:         fs,
			CurrentDir: currentDir,
			Defaulter:  defaults.NewDefaulter(projectData),
			Linters: []linters.Linter{
				linters.NewTeamLinter(),
				linters.NewRepoLinter(fs, currentDir, project.BranchResolver),
				linters.NewSecretsLinter(man.NewSecretValidator()),
				linters.NewTasksLinter(fs),
				linters.NewCfManifestLinter(manifest.ReadAndInterpolateManifest),
				linters.NewTriggerLinter(),
			},
			Renderer: pipeline.NewPipeline(manifest.ReadAndInterpolateManifest, fs),
		}

		pipelineConfig, lintResults := ctrl.Process()

		if lintResults.HasErrors() || lintResults.HasWarnings() {
			printErr(fmt.Errorf(lintResults.Error()))
			if lintResults.HasErrors() {
				os.Exit(1)
			}
		}

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
