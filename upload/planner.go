package upload

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
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
		return man, err
	}

	yamlString, err := filechecker.ReadFile(p.fs, halfpipeFilePath)
	if err != nil {
		return man, err
	}

	err = yaml.Unmarshal([]byte(yamlString), &man)
	if err != nil {
		return man, err
	}

	if man.Team == "" || man.Pipeline == "" {
		err = errors.New("'team' and 'pipeline' must be defined in '.halfpipe.io'")
	}

	return man, err
}

func (p planner) lintAndRender() (cmd Command, err error) {
	file, err := p.pipelineFile(p.fs)
	if err != nil {
		return cmd, err
	}

	path, err := p.pathResolver("halfpipe")
	if err != nil {
		return cmd, err
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{path},
		Stdout: file,
	}
	cmd.Printable = fmt.Sprintf("%s > %s", "halfpipe", file.Name())

	return cmd, err
}

func (p planner) statusAndLogin(concourseURL, team string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		err = ErrFlyNotInstalled(p.oSResolver())
		return cmd, err
	}

	cmd = Command{
		Cmd: exec.Cmd{
			Path:   path,
			Args:   []string{"fly", "-t", team, "status"},
			Stdout: ioutil.Discard,
		},
		ExecuteOnFailureFilter: func(outputFromPreviousCommand []byte) bool {
			knownErrors := []string{
				"unknown target",
				"Token is expired",
				"please login again",
			}

			for _, knownError := range knownErrors {
				if strings.Contains(string(outputFromPreviousCommand), knownError) {
					return true
				}
			}
			return false
		},
		ExecuteOnFailure: Plan{
			{
				Cmd: exec.Cmd{
					Path: path,
					Args: []string{"fly", "-t", team, "login", "-c", concourseURL, "-n", team},
				},
			},
		},
	}

	return cmd, err
}

func (p planner) uploadCmd(team, pipeline string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		err = ErrFlyNotInstalled(p.oSResolver())
		return cmd, err
	}

	cmd.Cmd = exec.Cmd{
		Path: path,
		Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds"},
	}

	if p.nonInteractive {
		cmd.Cmd.Args = append(cmd.Cmd.Args, "--non-interactive")
	}

	return cmd, err
}

func (p planner) Plan() (plan Plan, err error) {
	man, err := p.getHalfpipeManifest()
	if err != nil {
		return plan, err
	}

	if p.currentBranch != "master" && !p.nonInteractive {
		plan = append(plan, SecurityQuestion(man.Pipeline, p.currentBranch))
	}

	lintAndRenderCmd, err := p.lintAndRender()
	if err != nil {
		return plan, err
	}
	plan = append(plan, lintAndRenderCmd)

	concourseURL := config.ConcourseURL
	if p.envResolver("CONCOURSE_URL") != "" {
		concourseURL = p.envResolver("CONCOURSE_URL")
	}

	statusAndLoginCmd, err := p.statusAndLogin(concourseURL, man.Team)
	if err != nil {
		return plan, err
	}
	plan = append(plan, statusAndLoginCmd)

	uploadCmd, err := p.uploadCmd(man.Team, man.PipelineName())
	if err != nil {
		return plan, err
	}
	plan = append(plan, uploadCmd)

	return plan, err
}

func (p planner) Unpause() (plan Plan, err error) {
	man, err := p.getHalfpipeManifest()
	if err != nil {
		return plan, err
	}

	path, err := p.pathResolver("fly")
	if err != nil {
		return plan, err
	}

	plan = append(plan, Command{
		Cmd: exec.Cmd{
			Path: path,
			Args: []string{"fly", "-t", man.Team, "unpause-pipeline", "-p", man.PipelineName()},
		},
	})

	return plan, err
}
