package dependabot

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io/fs"
	"strings"
)

var ErrNotInGitRoot = errors.New("Must be executed in root of git repo")

type Walker interface {
	Walk() ([]string, error)
}

type walker struct {
	depth       int
	skipFolders []string
	fs          afero.Afero
}

func (w walker) gitExists() (err error) {
	if gitExists, err := w.fs.DirExists(".git"); err != nil || !gitExists {
		if err != nil {
			return err
		}
		if !gitExists {
			return ErrNotInGitRoot
		}
	}
	return nil
}

func (w walker) skipFolder(path string, skipFolders []string) (bool, string) {
	for _, skipFolder := range skipFolders {
		if strings.HasPrefix(path, skipFolder) {
			return true, skipFolder
		}
	}
	return false, ""
}

func (w walker) Walk() (paths []string, err error) {
	if err = w.gitExists(); err != nil {
		return
	}

	err = w.fs.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			logrus.Debug("Skipping '.git'")
			return fs.SkipDir
		}

		if skip, folder := w.skipFolder(path, w.skipFolders); skip {
			logrus.Debugf("Skipping '%s' because of skip '%s'", path, folder)
			return fs.SkipDir
		}

		if info.IsDir() && info.Name() == "node_modules" {
			logrus.Debug("Skipping 'node_modules'")
			return fs.SkipDir
		}

		if strings.Count(path, "/") > w.depth {
			logrus.Debugf("Skipping '%s' due to depth", path)
			return fs.SkipDir
		}

		if path != "." && !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	logrus.Debug("=========================")
	logrus.Debug("Found the following files")
	for _, path := range paths {
		logrus.Debug(path)
	}
	return
}

func NewWalker(fs afero.Afero, depth int, skipFolders []string) Walker {
	return walker{
		fs:          fs,
		depth:       depth,
		skipFolders: skipFolders,
	}
}
