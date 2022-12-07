package actions

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_tagWithCachePathHalfpipeIO(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "eu.gcr.io/halfpipe-io/image-name",
	}

	actual := tagWithCachePath(dockerPush)

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/image-name:${{ env.GIT_REVISION }}", actual)
}

func Test_tagWithCachePathHalfpipeIOAndTeam(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "eu.gcr.io/halfpipe-io/team/image-name",
	}

	actual := tagWithCachePath(dockerPush)

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/team/image-name:${{ env.GIT_REVISION }}", actual)
}

func Test_tagWithCachePathDockerHubRegistry(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "halfpipe/user",
	}

	actual := tagWithCachePath(dockerPush)

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/halfpipe/user:${{ env.GIT_REVISION }}", actual)
}
