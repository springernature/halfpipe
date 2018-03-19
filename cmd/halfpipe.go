package main

import (
	"fmt"
	"io"
	"os"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/secrets"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
)

func main() {
	var output string
	var err error

	switch {
	case invokedForHelp(os.Args):
		output, err = printHelp()
	case invokedForVersion(os.Args):
		output, err = printVersion()
	case invokedForSync(os.Args):
		err = syncBinary(os.Stdout)
	default:
		if err = checkVersion(); err != nil {
			break
		}
		output, err = lintAndRender()
	}

	fmt.Fprintln(os.Stdout, output) // #nosec

	if err != nil {
		fmt.Fprintln(os.Stderr, err) // #nosec
		os.Exit(1)
	}

}

func invokedForHelp(args []string) bool {
	return len(args) > 1 && (args[1] == "help" || args[1] == "-h" || args[1] == "-help" || args[1] == "--help")
}

func invokedForVersion(args []string) bool {
	return len(args) > 1 && (args[1] == "version" || args[1] == "-v" || args[1] == "-version" || args[1] == "--version")
}

func printHelp() (string, error) {
	version, err := config.GetVersion()
	return fmt.Sprintf(`Sup! Docs are at https://docs.halfpipe.io")
Current version is %s

Available commands are
  sync - updates the halfpipe cli to latest version 'halfpipe sync'
  help - this info
  version - version info
`, version), err
}

func printVersion() (string, error) {
	version, err := config.GetVersion()
	return version.String(), err
}

func invokedForSync(args []string) bool {
	return len(args) > 1 && args[1] == "sync"

}

func syncBinary(writer io.Writer) (err error) {
	currentVersion, err := config.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion, github.NewClient(nil).Repositories)
	err = syncer.Update(writer)
	return
}

func lintAndRender() (output string, err error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}

	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	project, err := defaults.NewProjectResolver(fs).Parse(currentDir)
	if err != nil {
		return
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
		Renderer: pipeline.Pipeline{},
	}

	pipelineConfig, lintResults := ctrl.Process()

	err = errors.New("")
	for _, result := range lintResults {
		err = errors.New(err.Error() + result.Error())
	}

	if lintResults.HasErrors() {
		return
	}

	output, renderError := pipeline.ToString(pipelineConfig)
	if renderError != nil {
		err = fmt.Errorf("%s\n%s", err, renderError)
	}

	return
}

func checkVersion() (err error) {
	currentVersion, err := config.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion, github.NewClient(nil).Repositories)
	err = syncer.Check()
	return
}
