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
	assert.Equal(t, "eu.gcr.io/halfpipe-io/engineering-enablement/o11ytool:0.2.11", run.Docker.Image, "wrong image")
	assert.Equal(t, `\upload-slos -i the-folder`, run.Script, "wrong script")
	assert.Equal(t, "the-name", run.Name, "wrong name")
	assert.Equal(t, manifest.Vars{"BOB": "BEN"}, run.Vars, "wrong vars")
}
