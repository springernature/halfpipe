package config

import "github.com/blang/semver"

var (
	// These fields will be populated in build
	// go build -ldflags "-X config.version=..."
	Version    string
	CompiledAt string
	GitCommit  string

	DocHost     string
	VaultPrefix string

	SlackWebhook = "Set your slack webhook here"

	DockerRegistry = "eu.gcr.io/halfpipe-io/"
)

var DevVersion = semver.Version{
	Major: 0,
	Minor: 0,
	Patch: 0,
	Pre:   []semver.PRVersion{{VersionStr: "DEV"}},
}

const HalfpipeFile = ".halfpipe.io"

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
