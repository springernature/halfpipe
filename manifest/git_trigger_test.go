package manifest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindsRepoNameWhenHTTP(t *testing.T) {
	git := GitTrigger{
		URI: "https://github.com/springernature/halfpipe.git",
	}
	assert.Equal(t, "halfpipe", git.GetRepoName())
}

func TestFindsRepoNameWhenSSH(t *testing.T) {
	git := GitTrigger{
		URI: "git@github.com:springernature/halfpipe.git",
	}
	assert.Equal(t, "halfpipe", git.GetRepoName())
}

func TestFindsRepoNameWithoutDotGit(t *testing.T) {
	git := GitTrigger{
		URI: "git@github.com:springernature/halfpipe",
	}
	assert.Equal(t, "halfpipe", git.GetRepoName())
}

func TestFindsOrgAndRepoName(t *testing.T) {
	t.Run("https", func(t *testing.T) {
		assert.Equal(t, "springernature/halfpipe", GitTrigger{
			URI: "https://github.com/springernature/halfpipe.git",
		}.GetOrgRepo())

		assert.Equal(t, "springernature/halfpipe", GitTrigger{
			URI: "https://github.com/springernature/halfpipe",
		}.GetOrgRepo())
	})

	t.Run("ssh", func(t *testing.T) {
		assert.Equal(t, "springernature/halfpipe", GitTrigger{
			URI: "git@github.com:springernature/halfpipe.git",
		}.GetOrgRepo())

		assert.Equal(t, "springernature/halfpipe", GitTrigger{
			URI: "git@github.com:springernature/halfpipe.git",
		}.GetOrgRepo())
	})
}
