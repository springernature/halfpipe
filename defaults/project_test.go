package defaults

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func testConfig() config {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	conf := NewConfig(fs)
	conf.LookPath = func(s string) (string, error) { return s, nil }
	conf.OriginURL = func() (string, error) { return "git@origin", nil }
	return conf
}

func TestErrorsIfGitNotFoundOnPath(t *testing.T) {
	conf := testConfig()
	conf.LookPath = func(string) (string, error) { return "", errors.New("dummy") }

	_, err := conf.Parse("/project/root")
	assert.Equal(t, ErrGitNotFound, err)
}

func TestErrorsIfNotInGitRepo(t *testing.T) {
	conf := testConfig()
	conf.OriginURL = func() (string, error) { return "", errors.New("dummy") }

	_, err := conf.Parse("/project/root")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestGetsGitOrigin(t *testing.T) {
	conf := testConfig()
	conf.Fs.MkdirAll("/project/root/.git", 0777)

	project, err := conf.Parse("/project/root")

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitUri)
}

func TestErrorsOutIfStartPathCannotBeRead(t *testing.T) {
	conf := testConfig()

	_, err := conf.Parse("/home/simon/src/repo")
	assert.Equal(t, ErrNotInRepo, err)
}

func TestBasePathWhenInGitRepo(t *testing.T) {
	conf := testConfig()
	conf.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	conf.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)

	assertBasePath(t, conf, "/home/simon/src/repo", "")
	assertBasePath(t, conf, "/home/simon/src/repo/sub1", "sub1")
	assertBasePath(t, conf, "/home/simon/src/repo/sub1/sub2", "sub1/sub2")
	assertBasePath(t, conf, "/home/simon/src/repo/sub1/sub2/sub3", "sub1/sub2/sub3")
}

func assertBasePath(t *testing.T, conf config, workingDir string, expectedBasePath string) {
	t.Helper()
	project, err := conf.Parse(workingDir)
	assert.Nil(t, err)
	assert.Equal(t, expectedBasePath, project.BasePath)
}

func TestErrorsOutIfWeReachRootWithoutFindingGit(t *testing.T) {
	conf := testConfig()
	conf.Fs.MkdirAll("/home/simon/src/repo/a/b/c", 0777)

	_, err := conf.Parse("/home/simon/src/repo/a/b/c")
	assert.Equal(t, ErrNotInRepo, err)
}
