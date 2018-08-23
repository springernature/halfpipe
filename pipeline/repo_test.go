package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersHttpGitResource(t *testing.T) {
	name := gitDir
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)

	man := manifest.Manifest{}
	man.Repo.URI = gitURI

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitURI,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersSshGitResource(t *testing.T) {
	name := gitDir
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.PrivateKey = privateKey

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":         gitURI,
					"private_key": privateKey,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithWatchesAndIgnores(t *testing.T) {
	name := gitDir
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.PrivateKey = privateKey

	watches := []string{"watch1", "watch2"}
	ignores := []string{"ignore1", "ignore2"}
	man.Repo.WatchedPaths = watches
	man.Repo.IgnoredPaths = ignores

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":          gitURI,
					"private_key":  privateKey,
					"paths":        watches,
					"ignore_paths": ignores,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersHttpGitResourceWithGitCrypt(t *testing.T) {
	name := gitDir
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	gitCrypt := "AABBFF66"

	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.GitCryptKey = gitCrypt

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":           gitURI,
					"git_crypt_key": gitCrypt,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithBranchIfSet(t *testing.T) {
	name := gitDir
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	branch := "master"

	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.Branch = branch

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":    gitURI,
					"branch": branch,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}
