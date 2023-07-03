package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlatformDefaultsToConcourseIfNotSet(t *testing.T) {
	expected := manifest.Manifest{Platform: "concourse"}
	assert.Equal(t, expected, newTopLevelDefaulter().Apply(manifest.Manifest{}))
}

func TestPlatformDoesNothingIfAlreadySet(t *testing.T) {
	expected := manifest.Manifest{Platform: "kehe"}
	assert.Equal(t, expected, newTopLevelDefaulter().Apply(manifest.Manifest{Platform: "kehe"}))
}

func TestPipelineIdDefaultsToPipelineNameIfNotSet(t *testing.T) {
	defaulted := newTopLevelDefaulter().Apply(manifest.Manifest{Pipeline: "kehe"})
	assert.Equal(t, "kehe", defaulted.PipelineId)
}

func TestPipelineIdDoesNothingIfAlreadySet(t *testing.T) {
	defaulted := newTopLevelDefaulter().Apply(manifest.Manifest{PipelineId: "id-provided"})
	assert.Equal(t, "id-provided", defaulted.PipelineId)
}
