package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLintsNothingIfNoDeployKateeTask(t *testing.T) {
	man := manifest.Manifest{}
	vl := VelaManifestLinter{Fs: afero.Afero{}}

	res := vl.Lint(man)

	assert.Len(t, res.Issues, 0)
}

func TestLintIfVelaFileDoesNotExist(t *testing.T) {
	velaManifestReader := afero.Afero{afero.NewMemMapFs()}
	vl := VelaManifestLinter{velaManifestReader}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployKatee{
				VelaManifest: "vela.yaml",
			},
		},
	}

	errs := vl.Lint(man)
	assertContainsError(t, errs.Issues, ErrFileNotFound)
}

func TestLintReturnsErrorIfVelaFileExistsButInvalid(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml", []byte("blah"), 0777)

	vl := VelaManifestLinter{fs}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployKatee{
				VelaManifest: "vela.yaml",
			},
		},
	}

	errs := vl.Lint(man)
	assertContainsError(t, errs.Issues, ErrFileInvalid)
}

func TestLintReturnsErrorIfEnvInKateeIsNotSetInHalfpipe(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
  components:
  - name: ${KATEE_APPLICATION_NAME}
    type: snstateless
    properties:
      image: ${KATEE_APPLICATION_IMAGE}
      env:
        - name: BLAH
          value: ${BLAH}
`), 0777)

	vl := VelaManifestLinter{fs}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployKatee{
				VelaManifest: "vela.yaml",
			},
		},
	}

	errs := vl.Lint(man)
	assertContainsError(t, errs.Issues, ErrVelaVariableMissing)
}

func TestLintReturnsNoErrorIfEnvInKateeIsSetInHalfpipe(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
  components:
  - name: ${KATEE_APPLICATION_NAME}
    type: snstateless
    properties:
      image: ${KATEE_APPLICATION_IMAGE}
      env:
        - name: haha
          value: ${BLAH}
`), 0777)

	vl := VelaManifestLinter{fs}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployKatee{
				VelaManifest: "vela.yaml",
				Vars: map[string]string{
					"BLAH": "Simon",
				},
			},
		},
	}

	errs := vl.Lint(man)
	assertNotContainsError(t, errs.Issues, ErrVelaVariableMissing)
}

func TestLintReturnsNoErrorIfEnvVarsInKateeAreBuildVersionOrGitRef(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
  components:
  - name: ${KATEE_APPLICATION_NAME}
    type: snstateless
    properties:
      image: ${KATEE_APPLICATION_IMAGE}
      env:
        - name: haha
          value: ${BUILD_VERSION}
`), 0777)

	vl := VelaManifestLinter{fs}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployKatee{
				VelaManifest: "vela.yaml",
				Vars: map[string]string{
					"BLAH": "Simon",
				},
			},
		},
	}

	errs := vl.Lint(man)

	assertNotContainsError(t, errs.Issues, ErrVelaVariableMissing)
}

func TestVelaManifestFileCanBeMarshalled(t *testing.T) {
	velaManifest, _ := unMarshallVelaManifest([]byte(
		`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
  components:
  - name: ${KATEE_APPLICATION_NAME}
    type: snstateless
    properties:
      image: ${KATEE_APPLICATION_IMAGE}
      env:
        - name: BLAH
          value: ${BLAH}
`))

	assert.Equal(t, "BLAH", velaManifest.Spec.Components[0].Properties.Env[0].Name)
}
