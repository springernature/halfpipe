package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineNameOnMaster(t *testing.T) {
	assert.Equal(t, "some-pipeline-name", actualName("some-pipeline-name", ""))
	assert.Equal(t, "some-pipeline-name", actualName("some-pipeline-name", "master"))
	assert.Equal(t, "some-pipeline-name", actualName("some-pipeline-name", "main"))
}

func TestPipelineNameOnBranch(t *testing.T) {
	assert.Equal(t, "some-pipeline-name-some-branch", actualName("some-pipeline-name", "some-branch"))
}

func TestPipelineNameShouldSanitizeDodgyCharactersInRepoAndBranchName(t *testing.T) {
	assert.Equal(t, "soME_pipeline-name99-some_branch", actualName(" soME$pipeline-name99 ", " some/branch "))
}

func actualName(repoName, branchName string) string {
	return Manifest{
		Pipeline: repoName,
		Triggers: TriggerList{
			GitTrigger{
				Branch: branchName,
			},
		},
	}.PipelineName()
}
