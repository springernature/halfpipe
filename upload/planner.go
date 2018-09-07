package upload

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
)

type PathResolver func(string) (string, error)
type PipelineFile func(fs afero.Afero) (afero.File, error)

type Targets struct {
	Targets map[string]interface{}
}

type Planner interface {
	Plan() (plan Plan, err error)
	Unpause() (plan Plan, err error)
	AskBranchSecurityQuestions(currentBranch string) (err error)
}

func NewPlanner(fs afero.Afero, pathResolver PathResolver, homedir string, stdout io.Writer, stderr io.Writer, stdin io.Reader, pipelineFile PipelineFile, nonInteractive bool) Planner {
	return planner{
		fs:             fs,
		pathResolver:   pathResolver,
		homedir:        homedir,
		stdout:         stdout,
		stderr:         stderr,
		stdin:          stdin,
		pipelineFile:   pipelineFile,
		nonInteractive: nonInteractive,
	}
}

type planner struct {
	fs             afero.Afero
	pathResolver   PathResolver
	homedir        string
	stdout         io.Writer
	stderr         io.Writer
	stdin          io.Reader
	pipelineFile   PipelineFile
	nonInteractive bool
}

func (p planner) getHalfpipeManifest() (man manifest.Manifest, err error) {
	yamlString, err := filechecker.ReadHalfpipeFiles(p.fs, config.HalfpipeFilenameOptions)
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

func (p planner) loginCommand(team string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		return
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{"fly", "-t", team, "login", "-c", "https://concourse.halfpipe.io", "-n", team},
		Stdout: p.stdout,
		Stderr: p.stderr,
		Stdin:  p.stdin,
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
		Stderr: p.stderr,
		Stdout: file,
	}
	cmd.Printable = fmt.Sprintf("%s > %s", "halfpipe", file.Name())

	return
}

func (p planner) uploadCmd(team, pipeline string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		return
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds"},
		Stdout: p.stdout,
		Stderr: p.stderr,
		Stdin:  p.stdin,
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

	lintAndRenderCmd, err := p.lintAndRender()
	if err != nil {
		return
	}
	plan = append(plan, lintAndRenderCmd)

	targets, err := p.getTargets()
	if err != nil {
		return
	}

	if _, ok := targets.Targets[man.Team]; !ok {
		cmd, loginError := p.loginCommand(man.Team)
		if loginError != nil {
			err = loginError
			return
		}

		plan = append(plan, cmd)
	}

	pipelineName := man.Pipeline

	if man.Repo.Branch != "" && man.Repo.Branch != "master" {
		pipelineName = man.Pipeline + "-" + man.Repo.Branch
	}

	uploadCmd, err := p.uploadCmd(man.Team, pipelineName)
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
			Path:   path,
			Args:   []string{"fly", "-t", man.Team, "unpause-pipeline", "-p", man.Pipeline},
			Stdout: p.stdout,
			Stderr: p.stderr,
			Stdin:  p.stdin,
		},
	})

	return
}

func (p planner) AskBranchSecurityQuestions(currentBranch string) (err error) {
	fmt.Fprintln(p.stdout)
	fmt.Fprintln(p.stdout, "WARNING! You are running halfpipe on a branch. WARNING!")
	fmt.Fprintln(p.stdout)
	if secErr := askSecurityQuestion("Have you made sure any Cloud Foundry manifests you are using in deploy-cf tasks have different app name and routes than on the master branch? And have you read the docs at https://docs.halfpipe.io/branches [y/N]: ", []string{"y", "yes", "Y", "Yes", "YES"}, p.stdout, p.stdin); secErr != nil {
		err = secErr
		return
	}

	man, manifestErr := p.getHalfpipeManifest()
	if manifestErr != nil {
		return
	}

	expectedAnswer := []string{fmt.Sprintf("%s-%s", man.Pipeline, currentBranch)}
	if secErr := askSecurityQuestion("What will be the name of the pipeline in concourse? (Hint, this is documented in the docs above): ", expectedAnswer, p.stdout, p.stdin); secErr != nil {
		err = secErr
		return
	}

	fmt.Fprintln(p.stdout)
	return
}

func askSecurityQuestion(question string, expectedAnswers []string, stdout io.Writer, stdin io.Reader) (err error) {
	fmt.Fprint(stdout, fmt.Sprintf("* %s", question))

	var input string
	fmt.Fscan(stdin, &input) // #nosec

	for _, expectedAnswer := range expectedAnswers {
		if input == expectedAnswer {
			return
		}
	}

	err = errors.New("Incorrect or empty response")
	return
}
