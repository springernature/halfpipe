package linters

//
//import (
//	"github.com/springernature/halfpipe/manifest"
//	"github.com/stretchr/testify/assert"
//	"testing"
//	"testing/fstest"
//)
//
//func TestLintsNothingIfNoDeployKateeTask(t *testing.T) {
//	man := manifest.Manifest{}
//	vl := VelaManifestLinter{fs: fstest.MapFS{}}
//
//	res := vl.Lint(man)
//
//	assert.Len(t, res.Errors, 0)
//}
//
//func TestLintIfVelaFileDoesNotExist(t *testing.T) {
//	velaManifestReader := fstest.MapFS{}
//	vl := VelaManifestLinter{velaManifestReader}
//
//	man := manifest.Manifest{
//		Tasks: []manifest.Task{
//			manifest.DeployKatee{
//				VelaManifest: "vela.yaml",
//			},
//		},
//	}
//
//	errs := vl.Lint(man)
//
//	assert.Len(t, errs.Errors, 1)
//}
//
//func TestLintReturnsNoErrorIfVelaFileExists(t *testing.T) {
//	velaManifestReader := fstest.MapFS{
//		"vela.yaml": {Data: []byte("blah")},
//	}
//	vl := VelaManifestLinter{velaManifestReader}
//
//	man := manifest.Manifest{
//		Tasks: []manifest.Task{
//			manifest.DeployKatee{
//				VelaManifest: "vela.yaml",
//			},
//		},
//	}
//
//	errs := vl.Lint(man)
//
//	assert.Len(t, errs.Errors, 0)
//}
//
//func TestLintReturnsErrorIfEnvNameDoesntCorrespondWithMan(t *testing.T) {
//
//}
//
//func TestVelaManifestFileCanBeMarshalled(t *testing.T) {
//	velaManifest := unMarshallVelaManifest([]byte(
//		`apiVersion: core.oam.dev/v1beta1
//kind: Application
//metadata:
// name: ${KATEE_APPLICATION_NAME}
// namespace: katee-${KATEE_TEAM}
//spec:
// //components:
//   - name: ${KATEE_APPLICATION_NAME}
//     type: snstateless
//     properties:
//       image: ${KATEE_APPLICATION_IMAGE}
//       env:
//         - name: BLAH
//           value: {{BLAH}}
//`))
//
//	assert.Equal(t, "BLAH", velaManifest.Spec.Components[0].Properties.Env[0].Value)
//}
