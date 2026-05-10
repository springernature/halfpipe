package shared

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

var DockerKateeCli = manifest.Docker{
	Image:    "eu.gcr.io/halfpipe-io/ee-run/docker/ee-katee-vela-cli:latest",
	Username: "oauth2accesstoken",
	Password: config.VaultSecrets.GARToken,
}
