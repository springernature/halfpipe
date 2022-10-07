package project

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tcnksm/go-gitconfig"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
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
	var pathRelativeToGit func(string, int) (basePath string, rootName string, gitRootPath string, err error)

	pathRelativeToGit = func(path string, depth int) (basePath string, rootName string, gitRootPath string, err error) {
		maxDepth := 50
		if !strings.Contains(path, string(filepath.Separator)) {
			return "", "", "", ErrNotInRepo
		}

		isGitWorkingTreeRoot, e := c.Fs.Exists(filepath.Join(path, ".git"))
		if e != nil {
			return "", "", "", e
		}

		switch {
		case isGitWorkingTreeRoot && path == workingDir:
			return "", filepath.Base(path), path, nil
		case isGitWorkingTreeRoot:
			basePath, err := filepath.Rel(path, workingDir)
			rootName := filepath.Base(path)
			return basePath, rootName, path, err
		case path == "/" || depth == maxDepth:
			return "", "", "", ErrNotInRepo
		default:
			return pathRelativeToGit(filepath.Join(path, ".."), depth+1)
		}
	}

	if _, e := c.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return p, err
	}

	basePath, rootName, gitRootPath, err := pathRelativeToGit(workingDir, 0)

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

	halfpipeFilePath, e := c.GetHalfpipeFileName(workingDir, halfpipeFilenameOptions)
	if errors.Is(e, ErrHalfpipeFileNotFound) && !ignoreMissingHalfpipeFile {
		return p, e
	}

	p.GitURI = origin
	p.GitRootPath = gitRootPath
	p.BasePath = basePath
	p.RootName = rootName
	p.HalfpipeFilePath = halfpipeFilePath
	return p, err
}

var ErrHalfpipeFileMultiple = errors.New("found multiple halfpipe manifests")
var ErrHalfpipeFileNotFound = errors.New("could not find halfpipe manifest")

func (c projectResolver) GetHalfpipeFileName(workingDir string, halfpipeFilenameOptions []string) (halfpipeFileName string, err error) {
	var foundPaths []string

	for _, p := range halfpipeFilenameOptions {
		joinedPath := path.Join(workingDir, p)
		if exists, fileNotExistErr := c.Fs.Exists(joinedPath); exists && fileNotExistErr == nil {
			foundPaths = append(foundPaths, p)
		}
	}

	if len(foundPaths) > 1 {
		err = fmt.Errorf("%w : %s", ErrHalfpipeFileMultiple, foundPaths)
		return halfpipeFileName, err
	}

	if len(foundPaths) == 0 {
		err = fmt.Errorf("%w : %s", ErrHalfpipeFileNotFound, halfpipeFilenameOptions)
		return halfpipeFileName, err
	}

	return foundPaths[0], nil
}
