package mapper

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestUploadSLOsMapper_ReturnsErrorWhenManifestCannotBeOpened(t *testing.T) {
	mapper := NewUploadSlosMapper()

	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{manifest.UploadSLOs{
		Folder: "the-folder",
		Name:   "the-name",
	}.SetVars(manifest.Vars{"BOB": "BEN"})}})
	assert.NoError(t, err)

	run := updated.Tasks[0].(manifest.Run)
	assert.Equal(t, "eu.gcr.io/halfpipe-io/engineering-enablement/o11ytool:0.2.10", run.Docker.Image, "wrong image")
	assert.Equal(t, `\upload-slos -i the-folder`, run.Script, "wrong script")
	assert.Equal(t, "the-name", run.Name, "wrong name")
	assert.Equal(t, manifest.Vars{"BOB": "BEN"}, run.Vars, "wrong vars")
}

//func TestUploadSLOsMapper_ParsesManifestAndPopulatesKateeManifest(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	velaContent := `kind: Application
//metadata:
//  name: "hello"
//  namespace: "default"
//spec:
//  components:
//    - name: my-app
//      type: snstateless
//      properties:
//        image: nginx:latest
//`
//	fs.WriteFile("vela.yaml", []byte(velaContent), 0644)
//
//	mapper := NewKateeMapper(fs)
//	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
//		manifest.DeployKatee{
//			VelaManifest: "vela.yaml",
//		},
//	}})
//
//	assert.NoError(t, err)
//	deployKatee := updated.Tasks[0].(manifest.DeployKatee)
//	assert.Equal(t, "Application", deployKatee.KateeManifest.Kind)
//	assert.Equal(t, "hello", deployKatee.KateeManifest.Metadata.Name)
//	assert.Equal(t, "default", deployKatee.KateeManifest.Metadata.Namespace)
//
//}
//
//func TestUploadSLOsMapper_HandlesParallelAndSequenceTasks(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	velaContent := `kind: Application
//spec:
//  components:
//    - name: nested-app
//      type: worker
//`
//	fs.WriteFile("vela.yaml", []byte(velaContent), 0644)
//
//	mapper := NewKateeMapper(fs)
//	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
//		manifest.Parallel{
//			Tasks: manifest.TaskList{
//				manifest.DeployKatee{
//					VelaManifest: "vela.yaml",
//				},
//				manifest.Sequence{
//					Tasks: manifest.TaskList{
//						manifest.DeployKatee{
//							VelaManifest: "vela.yaml",
//						},
//					},
//				},
//			},
//		},
//	}})
//
//	assert.NoError(t, err)
//
//	parallel := updated.Tasks[0].(manifest.Parallel)
//	deployKatee1 := parallel.Tasks[0].(manifest.DeployKatee)
//	assert.Equal(t, "Application", deployKatee1.KateeManifest.Kind)
//
//	sequence := parallel.Tasks[1].(manifest.Sequence)
//	deployKatee2 := sequence.Tasks[0].(manifest.DeployKatee)
//	assert.Equal(t, "Application", deployKatee2.KateeManifest.Kind)
//}
//
//func TestUploadSLOsMapper_ReturnsErrorForInvalidYAML(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	fs.WriteFile("vela.yaml", []byte("invalid: yaml: content:"), 0644)
//
//	mapper := NewKateeMapper(fs)
//	_, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
//		manifest.DeployKatee{
//			VelaManifest: "vela.yaml",
//		},
//	}})
//
//	assert.Error(t, err)
//}
