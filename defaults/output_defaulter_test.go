package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultsToConcourseIfNotSet(t *testing.T) {
	expected := manifest.Manifest{Platform: "concourse"}
	assert.Equal(t, expected, NewOutputDefaulter().Apply(manifest.Manifest{}))
}

func TestDoesNothingIfAlreadySet(t *testing.T) {
	expected := manifest.Manifest{Platform: "kehe"}
	assert.Equal(t, expected, NewOutputDefaulter().Apply(manifest.Manifest{Platform: "kehe"}))
}
