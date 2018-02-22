package path_to_git_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/afero"
	"path"
	"github.com/springernature/halfpipe/helpers/path_to_git"
)

func TestEmptyPathWhenStartPathHasGit(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	startPath := "/home/simon/src/repo"

	fs.Mkdir(startPath, 0777)
	fs.Mkdir(path.Join(startPath, ".git"), 0777)

	path, err := path_to_git.PathRelativeToGit(fs, startPath, -1)
	assert.Nil(t, err)
	assert.Equal(t, "", path)
}

func TestFindsPathWhenParentHaveGit(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	repoRoot := "/home/simon/src/repo"
	startPath := path.Join(repoRoot, "subfolder")

	fs.Mkdir(repoRoot, 0777)
	fs.Mkdir(startPath, 0777)
	fs.Mkdir(path.Join(repoRoot, ".git"), 0777)

	path, err := path_to_git.PathRelativeToGit(fs, startPath, -1)
	assert.Nil(t, err)
	assert.Equal(t, "subfolder", path)
}

func TestFindsPathWhenGrandParentHaveGit(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	repoRoot := "/home/simon/src/repo"
	startPath := path.Join(repoRoot, "a", "b", "c")
	fs.Mkdir(repoRoot, 0777)
	fs.Mkdir(startPath, 0777)
	fs.Mkdir(path.Join(repoRoot, ".git"), 0777)

	path, err := path_to_git.PathRelativeToGit(fs, startPath, -1)
	assert.Nil(t, err)
	assert.Equal(t, "a/b/c", path)
}

func TestErrorsOutIfWeReachRootWithoutFindingGit(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	repoRoot := "/home/simon/src/repo"
	startPath := path.Join(repoRoot, "a", "b", "c")
	fs.Mkdir(repoRoot, 0777)
	fs.Mkdir(startPath, 0777)

	_, err := path_to_git.PathRelativeToGit(fs, startPath, -1)
	assert.Error(t, err)
}

func TestErrorsOutIfWeHaveDoneMaxItterationsAndNotFoundGit(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	repoRoot := "/home/simon/src/repo"
	startPath := path.Join(repoRoot, "a", "b", "c")
	fs.Mkdir(repoRoot, 0777)
	fs.Mkdir(startPath, 0777)
	fs.Mkdir(path.Join(repoRoot, ".git"), 0777)

	_, err := path_to_git.PathRelativeToGit(fs, startPath, 2)
	assert.Error(t, err)
}


