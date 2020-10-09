package concourse

import (
	"testing"

	"regexp"

	con "github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: "repo.git",
			},
		},
	}

	actual, err := ToString(testPipeline().Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}

func TestToStringVersionComment(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: "repo.git",
			},
		},
	}
	con.Version = "0.0.1-yolo"

	actual, err := ToString(testPipeline().Render(man))

	assert.Nil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^#.*0\.0\.1-yolo.*`), actual)
}
