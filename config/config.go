package config

import (
	"github.com/blang/semver"
	"github.com/concourse/concourse/atc"
	"os"
)

// These fields will be populated in build
// go build -ldflags "-X config.version=..."
var (
	Version = "0.0.0-DEV"
)

var (
	Domain = getEnv("HALFPIPE_DOMAIN", "halfpipe.io")

	Project = getEnv("HALFPIPE_PROJECT", "halfpipe-io")

	DockerRegistry = "eu.gcr.io/" + Project + "/"

	DockerComposeImage = "halfpipe-docker-compose:stable"

	ConcourseURL = "https://concourse." + Domain

	CacheDirs = []atc.TaskCacheConfig{
		{Path: "../../../var/halfpipe/cache"},
		{Path: "../../../halfpipe-cache"}, // deprecated and should be removed after a while
	}

	DockerComposeCacheDirs = []string{
		"/var/halfpipe/cache",
		"/var/halfpipe/shared-cache",
	}
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

var HalfpipeFilenameOptions = []string{HalfpipeFile, HalfpipeFileWithYML, HalfpipeFileWithYAML}

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

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

const VersionBucket = "((halfpipe-semver.bucket))"
const VersionJSONKey = "((halfpipe-semver.private_key))"

const ArtifactsBucket = "((halfpipe-artifacts.bucket))"
const ArtifactsJSONKey = "((halfpipe-artifacts.private_key))"

const SlackWebhook = "((halfpipe-slack.webhook))"
