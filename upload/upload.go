package upload

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/concourse/fly/rc"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

type Targets map[rc.TargetName]rc.TargetProps
type FlyRcReader func() (Targets, error)
type LookPather func(string) (string, error)

type Upload interface {
	CreatePlan() (plan Plan, err error)
}

type upload struct {
	flyRcReader    FlyRcReader
	manifestReader manifest.ManifestReader
	fs             afero.Afero
	stdout         io.Writer
	stderr         io.Writer
	stdin          io.Reader
	currentDir     string
	lookPather     LookPather
}

func NewUpload(flyRcReader FlyRcReader, manifestReader manifest.ManifestReader, stdout io.Writer, stderr io.Writer, stdin io.Reader, lookPather LookPather, fs afero.Afero, currentDir string) Upload {
	return upload{
		flyRcReader:    flyRcReader,
		manifestReader: manifestReader,
		fs:             fs,
		stdout:         stdout,
		stderr:         stderr,
		stdin:          stdin,
		currentDir:     currentDir,
		lookPather:     lookPather,
	}
}

func (u upload) loginCommand(man manifest.Manifest, targets Targets) (plan Plan, err error) {
	if _, ok := targets[rc.TargetName(man.Team)]; !ok {
		path, errLookPather := u.lookPather("fly")
		if err != nil {
			err = errLookPather
			return
		}

		command := Command{
			Cmd: &exec.Cmd{
				Path:   path,
				Args:   []string{path, "-t", man.Team, "login", "-c", "https://concourse.halfpipe.io", "-n", man.Team},
				Stdin:  u.stdin,
				Stdout: u.stdout,
				Stderr: u.stderr,
			},
		}
		command.Printable = fmt.Sprintf("%s", strings.Join(command.Cmd.Args, " "))
		plan = append(plan, command)
	}
	return
}

func (u upload) CreatePlan() (plan Plan, err error) {
	man, err := u.manifestReader(u.currentDir, u.fs)
	if err != nil {
		return
	}

	if man.Team == "" || man.Pipeline == "" {
		err = errors.New("Top level fields 'team' and 'pipeline' must be set!")
		return
	}

	targets, err := u.flyRcReader()
	if err != nil {
		return
	}

	loginPlan, err := u.loginCommand(man, targets)
	if err != nil {
		return
	}

	plan = append(plan, loginPlan...)

	return
}
