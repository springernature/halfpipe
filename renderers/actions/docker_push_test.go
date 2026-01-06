package actions

import (
	"strings"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
	"github.com/stretchr/testify/assert"
)

func Test_tagWithCachePathHalfpipeIO(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "eu.gcr.io/halfpipe-io/image-name",
	}

	actual := shared.CachePath(dockerPush, ":${{ env.GIT_REVISION }}")

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/image-name:${{ env.GIT_REVISION }}", actual)
}

func Test_tagWithCachePathHalfpipeIOAndTeam(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "eu.gcr.io/halfpipe-io/team/image-name",
	}

	actual := shared.CachePath(dockerPush, ":${{ env.GIT_REVISION }}")

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/team/image-name:${{ env.GIT_REVISION }}", actual)
}

func Test_tagWithCachePathDockerHubRegistry(t *testing.T) {
	dockerPush := manifest.DockerPush{
		Image: "halfpipe/user",
	}

	actual := shared.CachePath(dockerPush, ":${{ env.GIT_REVISION }}")

	assert.Equal(t, "eu.gcr.io/halfpipe-io/cache/halfpipe/user:${{ env.GIT_REVISION }}", actual)
}

func Test_extractECRRepoName(t *testing.T) {
	t.Run("extracts repo name from ECR image", func(t *testing.T) {
		image := "744877006609.dkr.ecr.cn-northwest-1.amazonaws.com.cn/ee-run/testrepo"
		actual := extractECRRepoName(image)
		assert.Equal(t, "ee-run/testrepo", actual)
	})

	t.Run("extracts repo name from ECR image with tag", func(t *testing.T) {
		image := "744877006609.dkr.ecr.cn-northwest-1.amazonaws.com.cn/ee-run/testrepo:latest"
		actual := extractECRRepoName(image)
		assert.Equal(t, "ee-run/testrepo", actual)
	})

	t.Run("extracts repo name with multiple path segments", func(t *testing.T) {
		image := "744877006609.dkr.ecr.cn-northwest-1.amazonaws.com.cn/springernature/ee-run/helm-charts/myapp"
		actual := extractECRRepoName(image)
		assert.Equal(t, "springernature/ee-run/helm-charts/myapp", actual)
	})
}

func Test_dockerPushECRSteps(t *testing.T) {
	a := Actions{workingDir: "."}
	task := manifest.DockerPush{
		Image:          "744877006609.dkr.ecr.cn-northwest-1.amazonaws.com.cn/ee-run/testrepo",
		DockerfilePath: "Dockerfile",
	}
	man := manifest.Manifest{Platform: "actions", Team: "teamA"}

	steps := a.dockerPushSteps(task, man)

	// Should have 6 steps: AWS CLI, AWS credentials, ECR login, create repo, build/push, repository dispatch
	assert.Len(t, steps, 6)

	// First step should install AWS CLI
	assert.Equal(t, "Install AWS CLI", steps[0].Name)
	assert.Equal(t, ExternalActions.AWSCLI, steps[0].Uses)

	// Second step should be AWS credentials (using GitHub secrets)
	assert.Equal(t, "Configure AWS credentials", steps[1].Name)
	assert.Equal(t, ExternalActions.AWSCredentials, steps[1].Uses)
	assert.Equal(t, "${{ secrets.AWS_ACCESS_KEY_ID }}", steps[1].With["aws-access-key-id"])
	assert.Equal(t, "${{ secrets.AWS_SECRET_ACCESS_KEY }}", steps[1].With["aws-secret-access-key"])
	assert.Equal(t, "cn-northwest-1", steps[1].With["aws-region"])

	// Third step should be ECR login
	assert.Equal(t, "Login to Amazon ECR", steps[2].Name)
	assert.Equal(t, ExternalActions.AWSECRLogin, steps[2].Uses)
	assert.Equal(t, "login-ecr", steps[2].ID)

	// Fourth step should create ECR repo if not exists
	assert.Equal(t, "Create ECR repository (if not exists)", steps[3].Name)
	assert.True(t, strings.Contains(steps[3].Run, "aws ecr describe-repositories"))
	assert.True(t, strings.Contains(steps[3].Run, "aws ecr create-repository"))

	// Fifth step should build and push
	assert.Equal(t, "Build, tag, and push Docker image to ECR", steps[4].Name)
	assert.True(t, strings.Contains(steps[4].Run, "docker build"))
	assert.True(t, strings.Contains(steps[4].Run, "docker push"))

	// Sixth step should be repository dispatch
	assert.Equal(t, "Repository dispatch", steps[5].Name)
}
