package project

import (
	"bytes"
	"github.com/springernature/halfpipe/manifest"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func testProjectResolver() projectResolver {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pr := projectResolver{
		Fs:        fs,
		LookPath:  func(s string) (string, error) { return s, nil },
		OriginURL: func() (string, error) { return "git@origin", nil },
	}

	return pr
}

func TestErrors(t *testing.T) {
	t.Run("when git is not on the path", func(t *testing.T) {
		pr := testProjectResolver()
		pr.LookPath = func(string) (string, error) { return "", errors.New("dummy") }

		_, err := pr.Parse("/project/root")
		assert.Equal(t, ErrGitNotFound, err)
	})

	t.Run("when not in a git repo", func(t *testing.T) {
		pr := testProjectResolver()

		_, err := pr.Parse("/project/root")
		assert.Equal(t, ErrNotInRepo, err)
	})

	t.Run("when origin is not configured", func(t *testing.T) {
		pr := testProjectResolver()
		pr.OriginURL = func() (string, error) { return "", errors.New("dummy") }
		pr.Fs.MkdirAll("/project/root/.git", 0777)

		_, err := pr.Parse("/project/root")
		assert.Equal(t, ErrNoOriginConfigured, err)
	})

	t.Run("when start path cannot be reached", func(t *testing.T) {
		pr := testProjectResolver()

		_, err := pr.Parse("/home/simon/src/repo")
		assert.Equal(t, ErrNotInRepo, err)

	})

	t.Run("when we reach root without finding a repo", func(t *testing.T) {
		pr := testProjectResolver()
		pr.Fs.MkdirAll("/home/simon/src/repo/a/b/c", 0777)

		_, err := pr.Parse("/home/simon/src/repo/a/b/c")
		assert.Equal(t, ErrNotInRepo, err)
	})

	t.Run("when passed dodgy values", func(t *testing.T) {
		pr := testProjectResolver()

		paths := []string{"", "foo", "/..", ".."}

		for _, path := range paths {
			_, err := pr.Parse(path)
			assert.Equal(t, ErrNotInRepo, err)
		}
	})

	t.Run("when halfpipe file is missing", func(t *testing.T) {
		pr := testProjectResolver()
		pr.Fs.MkdirAll("/project/root/.git", 0777)

		project, err := pr.Parse("/project/root")

		assert.Nil(t, err)
		assert.Equal(t, "git@origin", project.GitURI)
	})

	t.Run("when halfpipe manifest is not a valid manifest", func(t *testing.T) {
		pr := testProjectResolver()
		pr.Fs.MkdirAll("/project/root/.git", 0777)
		pr.Fs.WriteFile("/project/root/.halfpipe.io", []byte("someRandomField: true"), 0777)

		_, err := pr.ShouldParseManifest().Parse("/project/root")

		assert.Error(t, err)
	})

	t.Run("when halfpipe manifest is not a valid manifest from stdin", func(t *testing.T) {
		pr := testProjectResolver()
		pr.Fs.MkdirAll("/project/root/.git", 0777)

		stdin := bytes.NewBufferString(`someRandomKey: true`)
		_, err := pr.LookForManifestOnStdIn(stdin).Parse("/project/root")

		assert.Error(t, err)
	})
}

func TestGetsGitData(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/project/root/.git", 0777)
	pr.Fs.WriteFile("/project/root/.halfpipe.io", []byte("team: myTeam"), 0777)

	project, err := pr.Parse("/project/root")

	assert.Nil(t, err)
	assert.Equal(t, "git@origin", project.GitURI)
}

func TestBasePathWhenInGitRepo(t *testing.T) {
	assertBasePath := func(t *testing.T, pr projectResolver, workingDir string, expectedBasePath string) {
		t.Helper()
		project, err := pr.Parse(workingDir)
		assert.Nil(t, err)
		assert.Equal(t, expectedBasePath, project.BasePath)
	}

	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/sub2/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/sub2/sub3/.halfpipe.io", []byte("team: myTeam"), 0777)

	assertBasePath(t, pr, "/home/simon/src/repo", "")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1", "sub1")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2", "sub1/sub2")
	assertBasePath(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "sub1/sub2/sub3")
}

func TestRootNameWhenInGitRepo(t *testing.T) {
	assertRootName := func(t *testing.T, pr projectResolver, workingDir string, expectedRootName string) {
		t.Helper()
		project, err := pr.Parse(workingDir)
		assert.Nil(t, err)
		assert.Equal(t, expectedRootName, project.RootName)
	}

	pr := testProjectResolver()
	pr.Fs.MkdirAll("/home/simon/src/repo/.git", 0777)
	pr.Fs.MkdirAll("/home/simon/src/repo/sub1/sub2/sub3", 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/sub2/.halfpipe.io", []byte("team: myTeam"), 0777)
	pr.Fs.WriteFile("/home/simon/src/repo/sub1/sub2/sub3/.halfpipe.io", []byte("team: myTeam"), 0777)

	assertRootName(t, pr, "/home/simon/src/repo", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2", "repo")
	assertRootName(t, pr, "/home/simon/src/repo/sub1/sub2/sub3", "repo")
}

func TestReadsFromStdin(t *testing.T) {
	pr := testProjectResolver()
	pr.Fs.MkdirAll("/project/root/.git", 0777)

	stdInManifest := `team: myTeam
pipeline: myPipeline`

	expectedManifest := manifest.Manifest{
		Team:     "myTeam",
		Pipeline: "myPipeline",
	}

	stdin := bytes.NewBufferString(stdInManifest)
	project, err := pr.LookForManifestOnStdIn(stdin).Parse("/project/root")

	assert.NoError(t, err)
	assert.Equal(t, expectedManifest, project.Manifest)
}
