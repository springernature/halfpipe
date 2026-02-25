package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCopyContainerImageDefault(t *testing.T) {

	t.Run("aws secrets", func(t *testing.T) {
		original := manifest.CopyContainerImage{}

		updated := copyContainerImageDefaulter(original, Concourse)
		assert.Equal(t, Concourse.AWSDocker.AccessKeyID, updated.AwsAccessKeyID)
		assert.Equal(t, Concourse.AWSDocker.SecretAccessKey, updated.AwsSecretAccessKey)
	})

}
