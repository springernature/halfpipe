package linters

import (
	"testing"

	"os"

	"fmt"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testCfManifestLinter() cfManifestLinter {
	return cfManifestLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}

	result := testCfManifestLinter().Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTask(t *testing.T) {
	data := []byte(`
---
applications:
- name: halfpipe-example-kotlin-dev
  instances: 1
`)

	testLinter := testCfManifestLinter()
	e := testLinter.Fs.WriteFile("manifest.yml", data, os.FileMode(0777))

	if e != nil {
		fmt.Print(e)
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := testLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTaskWithInvalidManifest(t *testing.T) {
	data := []byte(`
randomString
`)

	testLinter := testCfManifestLinter()
	testLinter.Fs.WriteFile("manifest.yml", data, os.FileMode(0777))

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := testLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Error(), "manifest.yml")
}

func TestOneCfDeployTaskWithNonExistingManifest(t *testing.T) {
	data := []byte(`
randomString
`)

	testLinter := testCfManifestLinter()
	testLinter.Fs.WriteFile("wrongManifest.yml", data, os.FileMode(0777))

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := testLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
}

func TestOneCfDeployTaskWithTwoApps(t *testing.T) {
	data := []byte(`
---
applications:
- name: halfpipe-example-kotlin-dev
  instances: 1
- name: halfpipe-example2
  instances: 2
`)
	testLinter := testCfManifestLinter()
	testLinter.Fs.WriteFile("manifest.yml", data, os.FileMode(755))

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := testLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertTooManyAppsError(t, "manifest.yml", result.Errors[0])
}

func TestTwoCfDeployTasksWithOneApp(t *testing.T) {
	data := []byte(`
---
applications:
- name: halfpipe-example-kotlin-dev
  instances: 1
- name: halfpipe-example-kotlin-dev
  instances: 1
`)
	testLinter := testCfManifestLinter()
	testLinter.Fs.WriteFile("manifest.yml", data, os.FileMode(755))

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := testLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertTooManyAppsError(t, "manifest.yml", result.Errors[0])
}
