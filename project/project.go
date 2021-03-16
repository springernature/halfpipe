package project

import (
	errors2 "github.com/springernature/halfpipe/linters/linterrors"
	"path/filepath"

	"github.com/springernature/halfpipe/linters/filechecker"
	//errors2 "github.com/springernature/halfpipe/linters/linterrors"

	"os/exec"

	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tcnksm/go-gitconfig"
)

type Data struct {
	BasePath         string
	RootName         string
	GitURI           string
	GitRootPath      string
	HalfpipeFilePath string
}

type Project interface {
	Parse(workingDir string, ignoreMissingHalfpipeFile bool, halfpipeFilenameOptions []string) (p Data, err error)
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
	ErrGitNotFound        = errors.New("looks like you don't have git installed")
	ErrNoOriginConfigured = errors.New("looks like you don't have a remote origin configured")
	ErrNotInRepo          = errors.New("looks like you are not executing halfpipe from within a git repo")
	ErrNoCommits          = errors.New("looks like you are executing halfpipe in a repo without commits, this is not supported")
)

func (c projectResolver) Parse(workingDir string, ignoreMissingHalfpipeFile bool, halfpipeFilenameOptions []string) (p Data, err error) {
	var pathRelativeToGit func(string) (basePath string, rootName string, gitRootPath string, err error)

	pathRelativeToGit = func(path string) (basePath string, rootName string, gitRootPath string, err error) {
		if !strings.Contains(path, string(filepath.Separator)) {
			return "", "", "", ErrNotInRepo
		}

		exists, e := c.Fs.DirExists(filepath.Join(path, ".git"))
		if e != nil {
			return "", "", "", e
		}

		switch {
		case exists && path == workingDir:
			return "", filepath.Base(path), path, nil
		case exists:
			basePath, err := filepath.Rel(path, workingDir)
			rootName := filepath.Base(path)
			return basePath, rootName, path, err
		case path == "/":
			return "", "", "", ErrNotInRepo
		default:
			return pathRelativeToGit(filepath.Join(path, ".."))
		}
	}

	if _, e := c.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return p, err
	}

	basePath, rootName, gitRootPath, err := pathRelativeToGit(workingDir)

	if err != nil {
		return p, err
	}
	// win -> unix path
	basePath = strings.Replace(basePath, `\`, "/", -1)

	origin, e := c.OriginURL()
	if e != nil {
		err = ErrNoOriginConfigured
		return p, err
	}

	halfpipeFilePath, e := filechecker.GetHalfpipeFileName(c.Fs, workingDir, halfpipeFilenameOptions)
	if e != nil {
		switch e.(type) {
		case errors2.MissingHalfpipeFileError:
			if !ignoreMissingHalfpipeFile {
				err = e
				return p, err
			}
		}
	}

	p.GitURI = origin
	p.GitRootPath = gitRootPath
	p.BasePath = basePath
	p.RootName = rootName
	p.HalfpipeFilePath = halfpipeFilePath
	return p, err
}
