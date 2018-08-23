package project

import (
	"path/filepath"

	"os/exec"

	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tcnksm/go-gitconfig"
	"bytes"
	"io/ioutil"
)

type Data struct {
	BasePath  string
	RootName  string
	GitURI    string
	GitBranch string
}

type GitBranchResolver func() (branch string, err error)

func BranchResolver() (branch string, err error) {
	if _, e := exec.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return
	}

	var stdout bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	if runErr := cmd.Run(); runErr != nil {
		err = runErr
		return
	}

	branch = strings.TrimSpace(stdout.String())
	if branch == "" {
		err = ErrBranchIsEmpty
		return
	}

	return
}

type Project interface {
	Parse(workingDir string) (p Data, err error)
}

type projectResolver struct {
	Fs             afero.Afero
	LookPath       func(string) (string, error)
	OriginURL      func() (string, error)
	BranchResolver GitBranchResolver
}

func NewProjectResolver(fs afero.Afero, branchResolver GitBranchResolver) projectResolver {
	return projectResolver{
		Fs:             fs,
		LookPath:       exec.LookPath,
		OriginURL:      gitconfig.OriginURL,
		BranchResolver: branchResolver,
	}
}

var (
	ErrGitNotFound        = errors.New("looks like you don't have git installed")
	ErrNoOriginConfigured = errors.New("looks like you don't have a remote origin configured")
	ErrNotInRepo          = errors.New("looks like you are not executing halfpipe from within a git repo")
	ErrBranchIsEmpty      = errors.New("looks like you are not on a branch?! This should never happen :o")
)

func (c projectResolver) Parse(workingDir string) (p Data, err error) {
	var pathRelativeToGit func(string) (basePath string, rootName string, err error)

	pathRelativeToGit = func(path string) (basePath string, rootName string, err error) {
		if !strings.Contains(path, string(filepath.Separator)) {
			return "", "", ErrNotInRepo
		}

		exists, e := c.Fs.DirExists(filepath.Join(path, ".git"))
		if e != nil {
			return "", "", e
		}

		switch {
		case exists && path == workingDir:
			return "", filepath.Base(path), nil
		case exists:
			basePath, err := filepath.Rel(path, workingDir)
			rootName := filepath.Base(path)
			return basePath, rootName, err
		case path == "/":
			return "", "", ErrNotInRepo
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
		err = ErrNoOriginConfigured
		return
	}

	basePath, rootName, err := pathRelativeToGit(workingDir)
	if err != nil {
		return
	}

	branch, err := c.BranchResolver()
	if err != nil {
		return
	}

	p.GitURI = origin
	p.BasePath = basePath
	p.RootName = rootName
	p.GitBranch = branch
	return
}
