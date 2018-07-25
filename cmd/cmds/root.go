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
	"github.com/springernature/halfpipe/linters/secrets"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/project"
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
	Short: `halfpipe is a tool to lint and render concoures pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkVersion(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fs := afero.Afero{Fs: afero.NewOsFs()}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		project, err := project.NewProjectResolver(fs).Parse(currentDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		ctrl := halfpipe.Controller{
			Fs:         fs,
			CurrentDir: currentDir,
			Defaulter:  defaults.NewDefaulter(project),
			Linters: []linters.Linter{
				linters.NewTeamLinter(),
				linters.NewRepoLinter(fs, currentDir),
				linters.NewSecretsLinter(config.VaultPrefix, secrets.NewSecretStore(fs)),
				linters.NewTasksLinter(fs),
				linters.NewCfManifestLinter(manifest.ReadAndMergeManifests),
				linters.NewArtifactsLinter(),
			},
			Renderer: pipeline.NewPipeline(manifest.ReadAndMergeManifests),
		}

		pipelineConfig, lintResults := ctrl.Process()

		if lintResults.HasErrors() || lintResults.HasWarnings() {
			fmt.Fprintln(os.Stderr, lintResults.Error())
			if lintResults.HasErrors() {
				os.Exit(1)
			}
		}

		pipelineString, renderError := pipeline.ToString(pipelineConfig)
		if renderError != nil {
			err = fmt.Errorf("%s\n%s", err, renderError)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println(pipelineString)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
