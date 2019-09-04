package project

import (
	"fmt"
	errors2 "github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"path/filepath"

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
	HalfpipeFilePath string
	Manifest         manifest.Manifest
}

type Project interface {
	Parse(workingDir string) (p Data, err error)
	ShouldParseManifest() Project
	ShouldFindManifestPath() Project
}

type projectResolver struct {
	Fs        afero.Afero
	LookPath  func(string) (string, error)
	OriginURL func() (string, error)

	findHalfpipeFileName bool
	readManifest         bool
}

func (c projectResolver) ShouldParseManifest() Project {
	newResolver := c
	newResolver.findHalfpipeFileName = true
	newResolver.readManifest = true
	return newResolver
}

func (c projectResolver) ShouldFindManifestPath() Project {
	newResolver := c
	newResolver.findHalfpipeFileName = true
	return newResolver
}

func NewProjectResolver(fs afero.Afero) Project {
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

func (c projectResolver) getGitOrigin() (origin string, err error) {
	if _, e := c.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return
	}

	origin, e := c.OriginURL()
	if e != nil {
		err = ErrNoOriginConfigured
		return
	}

	return
}

func (c projectResolver) pathRelativeToGit(workingDir string) (basePath string, rootName string, err error) {
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

	basePath, rootName, err = pathRelativeToGit(workingDir)
	if err != nil {
		return
	}

	// win -> unix path
	basePath = strings.Replace(basePath, `\`, "/", -1)
	return
}

func (c projectResolver) getHalfpipeFileName(workingDir string) (halfpipeFilePath string, err error) {
	halfpipeFilePath, e := filechecker.GetHalfpipeFileName(c.Fs, workingDir)
	if e != nil {
		if !(e == errors2.NewMissingHalfpipeFileError()) {
			err = e
			return
		}
	}

	return
}

func (c projectResolver) getManifest(workingDir, halfpipeFilePath string) (man manifest.Manifest, err error) {
	yaml, err := filechecker.ReadFile(c.Fs, path.Join(workingDir, halfpipeFilePath))
	if err != nil {
		return
	}
	//
	man, errs := manifest.Parse(yaml)
	if len(errs) != 0 {
		var errStrs []string
		for _, e := range errs {
			errStrs = append(errStrs, e.Error())
		}

		err = fmt.Errorf("could not parse manifest:\n%s", strings.Join(errStrs, "\n"))
		return
	}

	return
}

func (c projectResolver) Parse(workingDir string) (p Data, err error) {

	origin, err := c.getGitOrigin()
	if err != nil {
		return
	}
	p.GitURI = origin

	basePath, rootName, err := c.pathRelativeToGit(workingDir)
	if err != nil {
		return
	}
	p.BasePath = basePath
	p.RootName = rootName

	if c.findHalfpipeFileName {
		halfpipeFilePath, e := c.getHalfpipeFileName(workingDir)
		if e != nil {
			err = e
			return
		}
		p.HalfpipeFilePath = halfpipeFilePath

		if c.readManifest {
			man, e := c.getManifest(workingDir, halfpipeFilePath)
			if e != nil {
				err = e
				return
			}
			p.Manifest = man
		}
	}

	return
}
