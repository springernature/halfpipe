package upload

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
)

var ErrFlyNotInstalled = func(os string) error {
	return fmt.Errorf(`could not find the 'fly' binary. Please download it from here 'https://concourse.halfpipe.io/api/v1/cli?arch=amd64&platform=%s', make sure its called 'fly', is executable and put it on your path`, os)
}

type PathResolver func(string) (string, error)
type PipelineFile func(fs afero.Afero) (afero.File, error)
type OSResolver func() string
type EnvResolver func(envVar string) string

type Targets struct {
	Targets map[string]Target
}

type Target struct {
	API string
}

type Planner interface {
	Plan() (plan Plan, err error)
	Unpause() (plan Plan, err error)
}

func NewPlanner(fs afero.Afero, pathResolver PathResolver, homedir string, pipelineFile PipelineFile, nonInteractive bool, currentBranch string, osResolver OSResolver, envResolver EnvResolver, workingDir string) Planner {
	return planner{
		fs:             fs,
		pathResolver:   pathResolver,
		homedir:        homedir,
		pipelineFile:   pipelineFile,
		nonInteractive: nonInteractive,
		currentBranch:  currentBranch,
		oSResolver:     osResolver,
		envResolver:    envResolver,
		workingDir:     workingDir,
	}
}

type planner struct {
	fs             afero.Afero
	pathResolver   PathResolver
	homedir        string
	pipelineFile   PipelineFile
	nonInteractive bool
	currentBranch  string
	oSResolver     OSResolver
	envResolver    EnvResolver
	workingDir     string
}

func (p planner) getHalfpipeManifest() (man manifest.Manifest, err error) {
	halfpipeFilePath, err := filechecker.GetHalfpipeFileName(p.fs, p.workingDir)
	if err != nil {
		return
	}

	yamlString, err := filechecker.ReadFile(p.fs, halfpipeFilePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal([]byte(yamlString), &man)
	if err != nil {
		return
	}

	if man.Team == "" || man.Pipeline == "" {
		err = errors.New("'team' and 'pipeline' must be defined in '.halfpipe.io'")
	}

	return
}

func (p planner) getTargets() (targets Targets, err error) {
	path := filepath.Join(p.homedir, ".flyrc")
	exists, err := p.fs.Exists(path)
	if err != nil || !exists {
		return
	}

	bytes, err := p.fs.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(bytes, &targets)
	return
}

func (p planner) loginCommand(team string, host string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		err = ErrFlyNotInstalled(p.oSResolver())
		return
	}

	cmd.Cmd = exec.Cmd{
		Path: path,
		Args: []string{"fly", "-t", team, "login", "-c", host, "-n", team},
	}

	return
}

func (p planner) lintAndRender() (cmd Command, err error) {
	file, err := p.pipelineFile(p.fs)
	if err != nil {
		return
	}

	path, err := p.pathResolver("halfpipe")
	if err != nil {
		return
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{path},
		Stdout: file,
	}
	cmd.Printable = fmt.Sprintf("%s > %s", "halfpipe", file.Name())

	return
}

func (p planner) uploadCmd(team, pipeline string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		err = ErrFlyNotInstalled(p.oSResolver())
		return
	}

	cmd.Cmd = exec.Cmd{
		Path: path,
		Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds"},
	}

	if p.nonInteractive {
		cmd.Cmd.Args = append(cmd.Cmd.Args, "--non-interactive")
	}

	return
}

func (p planner) Plan() (plan Plan, err error) {
	man, err := p.getHalfpipeManifest()
	if err != nil {
		return
	}

	if p.currentBranch != "master" && !p.nonInteractive {
		plan = append(plan, SecurityQuestion(man.Pipeline, p.currentBranch))
	}

	lintAndRenderCmd, err := p.lintAndRender()
	if err != nil {
		return
	}
	plan = append(plan, lintAndRenderCmd)

	targets, err := p.getTargets()
	if err != nil {
		return
	}

	concourseOverrideHost := p.envResolver("CONCOURSE_HOST")
	if target, ok := targets.Targets[man.Team]; !ok || concourseOverrideHost != "" && target.API != concourseOverrideHost {
		host := config.ConcourseHost
		if concourseOverrideHost != "" {
			host = concourseOverrideHost
		}

		cmd, loginError := p.loginCommand(man.Team, host)
		if loginError != nil {
			err = loginError
			return
		}

		plan = append(plan, cmd)
	}

	uploadCmd, err := p.uploadCmd(man.Team, p.pipelineName(man))
	if err != nil {
		return
	}
	plan = append(plan, uploadCmd)

	return
}

func (p planner) Unpause() (plan Plan, err error) {
	man, err := p.getHalfpipeManifest()
	if err != nil {
		return
	}

	path, err := p.pathResolver("fly")
	if err != nil {
		return
	}

	plan = append(plan, Command{
		Cmd: exec.Cmd{
			Path: path,
			Args: []string{"fly", "-t", man.Team, "unpause-pipeline", "-p", p.pipelineName(man)},
		},
	})

	return
}

func (p planner) pipelineName(manifest manifest.Manifest) (pipelineName string) {
	pipelineName = manifest.Pipeline

	if manifest.Repo.Branch != "" && manifest.Repo.Branch != "master" {
		pipelineName = fmt.Sprintf("%s-%s", manifest.Pipeline, manifest.Repo.Branch)
	}

	return
}
