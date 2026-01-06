package defaults

import "github.com/springernature/halfpipe/manifest"

func dockerPushAWSDefaulter(original manifest.DockerPushAWS, man manifest.Manifest, defaults Defaults) (updated manifest.DockerPushAWS) {
	updated = original

	if updated.Region == "" {
		updated.Region = defaults.AWSDocker.Region
	}
	if updated.AccessKeyID == "" {
		updated.AccessKeyID = defaults.AWSDocker.AccessKeyID
	}
	if updated.SecretAccessKey == "" {
		updated.SecretAccessKey = defaults.AWSDocker.SecretAccessKey
	}

	if updated.DockerfilePath == "" {
		updated.DockerfilePath = defaults.Docker.FilePath
	}

	return updated
}
