package config

import "github.com/blang/semver"

var (
	// These fields will be populated in build
	// go build -ldflags "-X config.version=..."
	Version = "0.0.0-DEV"

	VaultPrefix string

	SlackWebhook = "Set your slack webhook here"

	DockerRegistry = "eu.gcr.io/halfpipe-io/"

	DockerComposeImage = "amidos/dcind:latest"

	ConsumerIntegrationTestImage = "eu.gcr.io/halfpipe-io/halfpipe-cdc-runner:latest"
)

var DevVersion = semver.Version{
	Major: 0,
	Minor: 0,
	Patch: 0,
	Pre:   []semver.PRVersion{{VersionStr: "DEV"}},
}

const HalfpipeFile = ".halfpipe.io"
const HalfpipeFileWithYML = ".halfpipe.io.yml"
const HalfpipeFileWithYAML = ".halfpipe.io.yaml"

var HalfpipeOptions = []string{HalfpipeFile, HalfpipeFileWithYML, HalfpipeFileWithYAML}

func GetVersion() (semver.Version, error) {
	if Version == "" {
		return DevVersion, nil
	}
	version, err := semver.Make(Version)
	if err != nil {
		return semver.Version{}, err
	}
	return version, nil
}
