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

func testProjectResolver(mockBranchResolver GitBranchResolver) projectResolver {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pr := NewProjectResolver(fs, mockBranchResolver)
	pr.LookPath = func(s string) (string, error) { return s, nil }
	pr.OriginURL = func() (string, error) { return "git@origin", nil }
	return pr
}

func TestErrorsIfGitNotFoundOnPath(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))
	pr.LookPath = func(string) (string, error) { return "", errors.New("dummy") }

	_, err := pr.Parse("/project/root")
	assert.Equal(t, ErrGitNotFound, err)
}

func TestErrorsIfNotInGitRepo(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))
	pr.OriginURL = func() (string, error) { return "", errors.New("dummy") }

	_, err := pr.Parse("/project/root")
	assert.Equal(t, ErrNoOriginConfigured, err)
}

func TestGetsGitData(t *testing.T) {
	branch := "my-cool-branch"
	pr := testProjectResolver(mockBranchResolver(branch, nil))
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	project, err := pr.Parse("/project/root")

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitURI)
	assert.Equal(t, branch, project.GitBranch)
}

func TestReturnsErrorFromBranchResolver(t *testing.T) {
	expectedError := errors.New("Meehp")
	pr := testProjectResolver(mockBranchResolver("", expectedError))
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	project, err := pr.Parse("/project/root")

	assert.Equal(t, expectedError, err)
	assert.Equal(t, Data{}, project)
}

func TestErrorsOutIfStartPathCannotBeRead(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))

	_, err := pr.Parse("/home/simon/src/repo")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestBasePathWhenInGitRepo(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)

	assertBasePath(t, pr, "/home/simon/src/repo", "")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1", "sub1")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2", "sub1/sub2")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "sub1/sub2/sub3")
}

func assertBasePath(t *testing.T, pr projectResolver, workingDir string, expectedBasePath string) {
	t.Helper()
	project, err := pr.Parse(workingDir)
	assert.Nil(t, err)
	assert.Equal(t, expectedBasePath, project.BasePath)
}

func TestRootNameWhenInGitRepo(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)

	assertRootName(t, pr, "/home/simon/src/repo", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "repo")
}

func assertRootName(t *testing.T, pr projectResolver, workingDir string, expectedRootName string) {
	t.Helper()
	project, err := pr.Parse(workingDir)
	assert.Nil(t, err)
	assert.Equal(t, expectedRootName, project.RootName)
}

func TestErrorsOutIfWeReachRootWithoutFindingGit(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))
	pr.Fs.MkdirAll("/home/simon/src/repo/a/b/c", 0777)

	_, err := pr.Parse("/home/simon/src/repo/a/b/c")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestErrorsOutIfPassedDodgyPathValue(t *testing.T) {
	pr := testProjectResolver(mockBranchResolver("master", nil))

	paths := []string{"", "foo", "/..", ".."}

	for _, path := range paths {
		_, err := pr.Parse(path)
		assert.Equal(t, ErrNotInRepo, err)
	}
}
