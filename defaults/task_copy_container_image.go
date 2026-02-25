package defaults

import (
	"github.com/springernature/halfpipe/manifest"
)

func copyContainerImageDefaulter(original manifest.CopyContainerImage, defaults Defaults) (updated manifest.CopyContainerImage) {
	updated = original
	if updated.AwsAccessKeyID == "" {
		updated.AwsAccessKeyID = defaults.AWSDocker.AccessKeyID
	}
	if updated.AwsSecretAccessKey == "" {
		updated.AwsSecretAccessKey = defaults.AWSDocker.SecretAccessKey
	}
	return updated
}
