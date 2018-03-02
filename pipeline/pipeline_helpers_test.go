package pipeline

import (
	"testing"

	"regexp"

	con "github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/parser"
	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	man := parser.Manifest{}
	man.Repo.Uri = "repo.git"

	actual, err := ToString(testPipeline().Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}

func TestToStringVersionComment(t *testing.T) {
	man := parser.Manifest{}
	man.Repo.Uri = "repo.git"
	con.Version = "0.0.1-yolo"

	actual, err := ToString(testPipeline().Render(man))

	assert.Nil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^#.*0\.0\.1-yolo.*`), actual)
}

func TestGeneratesUniqueNamesForJobsAndResources(t *testing.T) {
	manifest := parser.Manifest{
		Repo: parser.Repo{Uri: "https://github.com/springernature/halfpipe.git"},
		Tasks: []parser.Task{
			parser.Run{Script: "asd.sh"},
			parser.Run{Script: "asd.sh"},
			parser.Run{Script: "asd.sh"},
			parser.Run{Script: "fgh.sh"},
			parser.DeployCF{Org: "ee", Space: "dev"},
			parser.DeployCF{Org: "ee", Space: "dev"},
			parser.DeployCF{Org: "ee", Space: "dev"},
			parser.DockerPush{},
			parser.DockerPush{},
			parser.DockerPush{},
		},
	}
	config := testPipeline().Render(manifest)

	expectedJobNames := []string{
		"run asd.sh",
		"run asd.sh (1)",
		"run asd.sh (2)",
		"run fgh.sh",
		"deploy-cf",
		"deploy-cf (1)",
		"deploy-cf (2)",
		"docker-push",
		"docker-push (1)",
		"docker-push (2)",
	}

	expectedResourceNames := []string{
		"halfpipe",
		"CF ee-dev",
		"CF ee-dev (1)",
		"CF ee-dev (2)",
		"Docker Registry",
		"Docker Registry (1)",
		"Docker Registry (2)",
	}

	assert.Len(t, config.Jobs, len(expectedJobNames))
	assert.Len(t, config.Resources, len(expectedResourceNames))

	for i, name := range expectedJobNames {
		assert.Equal(t, name, config.Jobs[i].Name)
	}

	for i, name := range expectedResourceNames {
		assert.Equal(t, name, config.Resources[i].Name)
	}

}
