package project

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var mockBranchResolver = func(b string, e error) GitBranchResolver {
	return func() (branch string, err error) {
		return b, e
	}
}

func testProjectResolver() projectResolver {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pr := NewProjectResolver(fs)
	pr.LookPath = func(s string) (string, error) { return s, nil }
	pr.OriginURL = func() (string, error) { return "git@origin", nil }
	return pr
}

func TestErrorsIfGitNotFoundOnPath(t *testing.T) {
	pr := testProjectResolver()
	pr.LookPath = func(string) (string, error) { return "", errors.New("dummy") }

	_, err := pr.Parse("/project/root", false)
	assert.Equal(t, ErrGitNotFound, err)
}

func TestErrorsIfNotInGitRepo(t *testing.T) {
	pr := testProjectResolver()

	_, err := pr.Parse("/project/root", false)
	assert.Equal(t, ErrNotInRepo, err)
}

func TestErrorsIfNoGitOriginConfigured(t *testing.T) {
	pr := testProjectResolver()
	pr.OriginURL = func() (string, error) { return "", errors.New("dummy") }
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	_, err := pr.Parse("/project/root", false)
	assert.Equal(t, ErrNoOriginConfigured, err)
}

func TestGetsGitData(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/project/root/.git", 0777)
	pr.Fs.Create("/project/root/.halfpipe.io")

	project, err := pr.Parse("/project/root", false)

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitURI)
}

func TestErrorsOutIfStartPathCannotBeRead(t *testing.T) {
	pr := testProjectResolver()

	_, err := pr.Parse("/home/simon/src/repo", false)
	assert.Equal(t, ErrNotInRepo, err)
}

func TestBasePathWhenInGitRepo(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)
	pr.Fs.Create("/home/simon/src/repo/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/sub2/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/sub2/sub3/.halfpipe.io")

	assertBasePath(t, pr, "/home/simon/src/repo", "")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1", "sub1")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2", "sub1/sub2")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "sub1/sub2/sub3")
}

func assertBasePath(t *testing.T, pr projectResolver, workingDir string, expectedBasePath string) {
	t.Helper()
	project, err := pr.Parse(workingDir, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBasePath, project.BasePath)
}

func TestRootNameWhenInGitRepo(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)
	pr.Fs.Create("/home/simon/src/repo/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/sub2/.halfpipe.io")
	pr.Fs.Create("/home/simon/src/repo/sub1/sub2/sub3/.halfpipe.io")

	assertRootName(t, pr, "/home/simon/src/repo", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "repo")
}

func assertRootName(t *testing.T, pr projectResolver, workingDir string, expectedRootName string) {
	t.Helper()
	project, err := pr.Parse(workingDir, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedRootName, project.RootName)
}

func TestErrorsOutIfWeReachRootWithoutFindingGit(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/a/b/c", 0777)

	_, err := pr.Parse("/home/simon/src/repo/a/b/c", false)
	assert.Equal(t, ErrNotInRepo, err)
}

func TestErrorsOutIfPassedDodgyPathValue(t *testing.T) {
	pr := testProjectResolver()

	paths := []string{"", "foo", "/..", ".."}

	for _, path := range paths {
		_, err := pr.Parse(path, false)
		assert.Equal(t, ErrNotInRepo, err)
	}
}

func TestDoesntErrorOutIfHalfpipeFileIsMissing(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	project, err := pr.Parse("/project/root", true)

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitURI)
}
