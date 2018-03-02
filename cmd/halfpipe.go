package main

import (
	"fmt"
	"os"
	"syscall"

	"io"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/secret_resolver"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
)

func main() {
	var output string
	var err error

	switch {
	case invokedForHelp(os.Args):
		output, err = printHelp()
	case invokedForSync(os.Args):
		err = syncBinary(os.Stdout)
	default:
		if err = checkVersion(); err != nil {
			break
		}
		output, err = lintAndRender()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		syscall.Exit(1)
	}
	fmt.Fprintln(os.Stdout, output)
}

func invokedForHelp(args []string) bool {
	return len(args) > 1 && (args[1] == "-h" || args[1] == "-help" || args[1] == "--help")
}

func printHelp() (string, error) {
	version, err := config.GetVersion()
	return fmt.Sprintf(`Sup! Docs are at https://docs.halfpipe.io")
Current version is %s
Available commands are
\tsync - updates the halfpipe cli to latest version 'halfpipe sync'
`, version), err
}

func invokedForSync(args []string) bool {
	return len(args) > 1 && args[1] == "sync"

}

func syncBinary(writer io.Writer) (err error) {
	currentVersion, err := config.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion)
	err = syncer.Update(writer)
	return
}

func lintAndRender() (output string, err error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}

	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	ctrl := halfpipe.Controller{
		Fs:         fs,
		CurrentDir: currentDir,
		Linters: []linters.Linter{
			linters.TeamLinter{},
			linters.RepoLinter{Fs: fs},
			linters.NewSecretsLinter(secret_resolver.NewConcourseResolver(config.VaultPrefix, secret_resolver.NewSecretResolver(fs))),
			linters.TaskLinter{Fs: fs},
			linters.ArtifactsLinter{},
		},
		Renderer:  pipeline.Pipeline{},
		Defaulter: defaults.DefaultValues.Update,
	}

	pipelineConfig, lintResults := ctrl.Process()

	err = errors.New("")
	if lintResults.HasErrors() {
		for _, e := range lintResults {
			err = errors.New(err.Error() + e.Error())
		}
		return
	}

	output, err = pipeline.ToString(pipelineConfig)
	return
}

func checkVersion() (err error) {
	currentVersion, err := config.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion)
	err = syncer.Check()
	return
}
