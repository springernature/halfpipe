package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultsToConcourseIfNotSet(t *testing.T) {
	expected := manifest.Manifest{Output: "concourse"}
	assert.Equal(t, expected, NewOutputDefaulter().Apply(manifest.Manifest{}))
}

func TestDoesNothingIfAlreadySet(t *testing.T) {
	expected := manifest.Manifest{Output: "kehe"}
	assert.Equal(t, expected, NewOutputDefaulter().Apply(manifest.Manifest{Output: "kehe"}))
}
