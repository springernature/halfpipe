package main

import (
	"fmt"
	"io"
	"os"

	"errors"
	"os/exec"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/concourse/fly/rc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/secrets"
	halfpipeManifest "github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/upload"
)

func main() {
	var output string
	var err error

	switch {
	case invokedForHelp(os.Args):
		output, err = printHelp()
	case invokedForInit(os.Args):
		output, err = generateSampleHalfpipeFile()
	case invokedForVersion(os.Args):
		output, err = printVersion()
	case invokedForSync(os.Args):
		err = syncBinary(os.Stdout)
	case invokedForUpload(os.Args):
		err = renderAndUpload()
		if err == nil {
			output = " "
		}
	default:
		if err = checkVersion(); err != nil {
			break
		}
		output, err = lintAndRender()
	}

	if output != "" {
		fmt.Fprintln(os.Stdout, output) // #nosec
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err) // #nosec
	}

	if output == "" {
		os.Exit(1)
	}
}

func invokedForHelp(args []string) bool {
	return len(args) > 1 && (args[1] == "help" || args[1] == "-h" || args[1] == "-help" || args[1] == "--help")
}

func invokedForInit(args []string) bool {
	return len(args) > 1 && (args[1] == "init")
}

func invokedForVersion(args []string) bool {
	return len(args) > 1 && (args[1] == "version" || args[1] == "-v" || args[1] == "-version" || args[1] == "--version")
}

func invokedForUpload(args []string) bool {
	return len(args) > 1 && args[1] == "upload"
}

func printHelp() (string, error) {
	version, err := config.GetVersion()
	return fmt.Sprintf(`Sup! Docs are at https://docs.halfpipe.io")
Current version is %s

Available commands are
  init - creates a sample .halfpipe.io file in the current directory
  sync - updates the halfpipe cli to latest version 'halfpipe sync'
  renderAndUpload - lint, renders and uploads the pipeline to concourse
  help - this info
  version - version info
`, version), err
}

func generateSampleHalfpipeFile() (output string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	err = halfpipeManifest.NewSampleGenerator(afero.Afero{Fs: afero.NewOsFs()}).Generate()
	if err != nil {
		return
	}
	output = fmt.Sprintf("Generated sample configuration at %s/.halfpipe.io", currentDir)

	return
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

	syncer := sync.NewSyncer(currentVersion, sync.ResolveLatestVersionFromArtifactory)
	err = syncer.Update(writer)
	return
}

func renderAndUpload() (err error) {
	targetsReader := func() (targets upload.Targets, err error) {
		details, err := rc.LoadTargets()
		if err != nil {
			return
		}

		targets = details.Targets
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	uploader := upload.NewUpload(targetsReader,
		halfpipeManifest.ReadManifest,
		os.Stdout,
		os.Stderr,
		os.Stdin,
		exec.LookPath,
		afero.Afero{Fs: afero.NewOsFs()},
		currentDir)

	p, err := uploader.CreatePlan()
	if err != nil {
		return err
	}

	p.PrintPlan()
	fmt.Print("Are you sure? [y/N]: ")
	var input string
	fmt.Scanln(&input)
	if input != "y" {
		return errors.New("Aborted upload")
	}

	err = p.Execute()
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
		Fs:             fs,
		CurrentDir:     currentDir,
		Defaulter:      defaults.NewDefaulter(project),
		ManifestReader: halfpipeManifest.ReadManifest,
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
		err = lintResults
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

	syncer := sync.NewSyncer(currentVersion, sync.ResolveLatestVersionFromArtifactory)
	err = syncer.Check()
	return
}
