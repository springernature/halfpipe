package pipeline

import (
	"testing"
	"github.com/concourse/atc"
	"fmt"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func TestRendersHttpGitResource(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)

	manifest := model.Manifest{
		Repo: model.Repo{
			Uri: gitUri,
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitUri,
				},
			},
		},
	}
	assert.Equal(t, expected, Pipeline{}.Render(manifest))
}

func TestRendersSshGitResource(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	private_key := "blurgh"

	manifest := model.Manifest{
		Repo: model.Repo{
			Uri:        gitUri,
			PrivateKey: private_key,
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitUri,
					"private_key": private_key,
				},
			},
		},
	}
	assert.Equal(t, expected, Pipeline{}.Render(manifest))
}
