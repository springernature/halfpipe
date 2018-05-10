package defaults

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

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

	_, err := pr.Parse("/project/root")
	assert.Equal(t, ErrGitNotFound, err)
}

func TestErrorsIfNotInGitRepo(t *testing.T) {
	pr := testProjectResolver()
	pr.OriginURL = func() (string, error) { return "", errors.New("dummy") }

	_, err := pr.Parse("/project/root")
	assert.Equal(t, ErrNoOriginConfigured, err)
}

func TestGetsGitOrigin(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	project, err := pr.Parse("/project/root")

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitURI)
}

func TestErrorsOutIfStartPathCannotBeRead(t *testing.T) {
	pr := testProjectResolver()

	_, err := pr.Parse("/home/simon/src/repo")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestBasePathWhenInGitRepo(t *testing.T) {
	pr := testProjectResolver()
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

func TestErrorsOutIfWeReachRootWithoutFindingGit(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/a/b/c", 0777)

	_, err := pr.Parse("/home/simon/src/repo/a/b/c")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestErrorsOutIfPassedDodgyPathValue(t *testing.T) {
	pr := testProjectResolver()

	paths := []string{"", "foo", "/..", ".."}

	for _, path := range paths {
		_, err := pr.Parse(path)
		assert.Equal(t, ErrNotInRepo, err)
	}
}
