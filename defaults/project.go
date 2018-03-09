package defaults

import (
	"path/filepath"

	"os/exec"

	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tcnksm/go-gitconfig"
)

type Project struct {
	BasePath string
	GitURI   string
}

type projectResolver struct {
	Fs        afero.Afero
	LookPath  func(string) (string, error)
	OriginURL func() (string, error)
}

func NewProjectResolver(fs afero.Afero) projectResolver {
	return projectResolver{
		Fs:        fs,
		LookPath:  exec.LookPath,
		OriginURL: gitconfig.OriginURL,
	}
}

var (
	ErrGitNotFound = errors.New("looks like you don't have git installed")
	ErrNotInRepo   = errors.New("looks like you are not executing halfpipe from within a git repo")
)

func (c projectResolver) Parse(workingDir string) (p Project, err error) {
	var pathRelativeToGit func(string) (string, error)

	pathRelativeToGit = func(path string) (string, error) {
		if !strings.Contains(path, string(filepath.Separator)) {
			return "", ErrNotInRepo
		}

		exists, err := c.Fs.DirExists(filepath.Join(path, ".git"))
		if err != nil {
			return "", err
		}

		switch {
		case exists && path == workingDir:
			return "", nil
		case exists:
			return filepath.Rel(path, workingDir)
		case path == "/":
			return "", ErrNotInRepo
		default:
			return pathRelativeToGit(filepath.Join(path, ".."))
		}
	}

	if _, e := c.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return
	}

	origin, e := c.OriginURL()
	if e != nil {
		err = ErrNotInRepo
		return
	}

	basePath, err := pathRelativeToGit(workingDir)
	if err != nil {
		return
	}

	p.GitURI = origin
	p.BasePath = basePath
	return
}
