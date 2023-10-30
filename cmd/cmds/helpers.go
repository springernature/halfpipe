package cmds

import (
	"fmt"
	"github.com/springernature/halfpipe/renderers/concourse"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/renderers/actions"
	"github.com/springernature/halfpipe/sync"
	"github.com/tcnksm/go-gitconfig"
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

func formatInput(input string) []string {
	halfpipeFilenameOptions := config.HalfpipeFilenameOptions
	if Input == "" {
		return halfpipeFilenameOptions
	}
	if strings.Contains(input, string(os.PathSeparator)) {
		fmt.Printf("Input file '%s' must be in current directory\n", input)
		os.Exit(1)
	}
	return []string{input}

}

func getManifest(fs afero.Afero, currentDir, halfpipeFilePath string) (man manifest.Manifest, errors []error) {
	yaml, err := linters.ReadFile(fs, path.Join(currentDir, halfpipeFilePath))
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

func outputLintResults(lintResults linters.LintResults) {
	if lintResults.HasWarnings() && !lintResults.HasErrors() && !Quiet {
		printErr(lintResults)
		return
	}

	if lintResults.HasErrors() {
		printErr(lintResults)
		os.Exit(1)
	}
}

func renderResponse(r halfpipe.Response, filePath string) {
	outputLintResults(r.LintResults)

	outputYaml := fmt.Sprintf("# Generated using halfpipe cli version %s\n%s", config.Version, r.ConfigYaml)

	if filePath == "" {
		fmt.Println(outputYaml)
		return
	}

	if !Quiet {
		fileType := "Concourse pipeline"
		if r.Platform.IsActions() {
			fileType = "GitHub Actions workflow"
		}
		fmt.Printf("Writing %s to %s\n", fileType, filePath)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		printErr(err)
		return
	}
	if err := os.WriteFile(filePath, []byte(outputYaml), 0644); err != nil {
		printErr(err)
		return
	}
}

func createDefaulter(projectData project.Data, renderer halfpipe.Renderer) defaults.Defaults {
	switch renderer.(type) {
	case actions.Actions:
		return defaults.New(defaults.Actions, projectData)
	default:
		return defaults.New(defaults.Concourse, projectData)
	}
}

func createController(projectData project.Data, fs afero.Afero, currentDir string, renderer halfpipe.Renderer) halfpipe.Controller {
	return halfpipe.NewController(
		createDefaulter(projectData, renderer),
		mapper.New(),
		[]linters.Linter{
			linters.NewTopLevelLinter(),
			linters.NewTriggersLinter(fs, currentDir, project.BranchResolver, gitconfig.OriginURL),
			linters.NewSecretsLinter(manifest.NewSecretValidator()),
			linters.NewTasksLinter(fs, runtime.GOOS),
			linters.NewFeatureToggleLinter(manifest.AvailableFeatureToggles),
			linters.NewActionsLinter(),
		},
		renderer,
	)

}

func getManifestAndController(halfpipeFilenameOptions []string, renderer halfpipe.Renderer) (manifest.Manifest, halfpipe.Controller) {
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

	projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false, halfpipeFilenameOptions)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	man, manErrors := getManifest(fs, currentDir, projectData.HalfpipeFilePath)
	if len(manErrors) > 0 {
		outputLintResults(linters.LintResults{linters.NewLintResult("Halfpipe Manifest", "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/", manErrors)})
	}

	if renderer == nil {
		if man.Platform.IsActions() {
			renderer = actions.NewActions(projectData.GitURI, projectData.HalfpipeFilePath)
		} else {
			renderer = concourse.NewPipeline(projectData.HalfpipeFilePath)
		}
	}
	controller := createController(projectData, fs, currentDir, renderer)

	return man, controller
}
