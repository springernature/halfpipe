package config

import (
	"github.com/blang/semver"
	"github.com/concourse/atc"
)

var (
	// These fields will be populated in build
	// go build -ldflags "-X config.version=..."
	Version = "0.0.0-DEV"

	SlackWebhook = "Set your slack webhook here"

	DockerRegistry = "eu.gcr.io/halfpipe-io/"

	DockerComposeImage = "eu.gcr.io/halfpipe-io/halfpipe-docker-compose:stable"

	PrometheusGatewayURL = "prometheus-pushgateway:9091"

	ConcourseHost = "https://concourse.halfpipe.io"

	CacheDirs = []atc.CacheConfig{
		{Path: "../../../halfpipe-cache"},
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

const VersionBucket = "halfpipe-io-semver"
const VersionJsonKey = "((gcr.private_key))"
