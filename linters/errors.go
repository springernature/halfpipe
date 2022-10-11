package linters

import (
	"errors"
	"fmt"
)

var (
	ErrMissingField    = newError("required field missing")
	NewErrMissingField = func(field string) Error { return ErrMissingField.WithValue(field) }

	ErrInvalidField    = newError("invalid field")
	NewErrInvalidField = func(field string, reason string) Error { return ErrInvalidField.WithValue(field).WithValue(reason) }

	ErrFileNotFound      = newError("file not found")
	ErrFileCannotRead    = newError("file cannot be read")
	ErrFileNotAFile      = newError("not a file")
	ErrFileEmpty         = newError("file is empty")
	ErrFileNotExecutable = newError("file is not executable")
	ErrFileInvalid       = newError("file is invalid")

	ErrCFMissingRoutes        = newError("cf application must have at least one route")
	ErrCFMissingName          = newError("cf application missing 'name'")
	ErrCFRoutesAndNoRoute     = newError("cf application cannot have both 'routes' and 'no-route'")
	ErrCFNoRouteHealthcheck   = newError("cf application with 'no-route: true' requires 'health-check-type: process'")
	ErrCFRouteScheme          = newError("cf application route must not start with http(s)://")
	ErrCFRouteMissing         = newError("cf application routes must contain sso_route")
	ErrCFMultipleApps         = newError("cf manifest must have exactly 1 application")
	ErrCFBuildpackUnversioned = newError("buildpack specified without version so the latest will be used on each deploy")
	ErrCFBuildpackMissing     = newError("buildpack missing. Cloud Foundry will try to detect which system buildpack to use. Please see <https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#buildpack>")
	ErrCFBuildpackDeprecated  = newError("'buildpack' is deprecated in favour of 'buildpacks'. Please see <http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated>")
	ErrCFArtifactAndDocker    = newError("cannot combine 'deploy_artifact' in the halfpipe task and 'docker' in the cf manifest")
	ErrCFFromArtifact         = newError("this file must be saved as an artifact in a previous task")
	ErrCFPrePromoteArtifact   = newError("cannot have pre promote tasks with CF manifest restored from artifact")

	ErrUnsupportedRegistry  = newError("image must be from halfpipe registry. Please see <https://ee.public.springernature.app/rel-eng/docker-registry/>")
	ErrDockerPushTag        = newError("the field 'tag' is no longer used and is safe to delete")
	ErrDockerComposeVersion = newError("the docker-compose file version used is deprecated. All services must be under the 'services' key and 'Version' must be '2' or higher. Please see <https://docs.docker.com/compose/compose-file/compose-versioning/#versioning>")
	ErrMultipleTriggers     = newError("cannot have multiple triggers of this type")

	ErrVelaVariableMissing = newError("vela manifest variable is not specified in halfpipe manifest")

	ErrUnsupportedManualTrigger   = newError("manual_trigger on individual tasks is not supported in GitHub Actions. It is supported at the workflow level in git trigger options")
	ErrUnsupportedRolling         = newError("cf rolling deploys are not supported in GitHub Actions")
	ErrDockerTriggerLoop          = newError("cannot push docker image that is also a trigger as it will create a loop")
	ErrUnsupportedCovenant        = newError("covenant is not supported in GitHub Actions")
	ErrUnsupportedGitPrivateKey   = newError("git private_key is not supported in GitHub Actions")
	ErrUnsupportedGitUri          = newError("git uri is not supported in GitHub Actions")
	ErrUnsupportedPipelineTrigger = newError("pipeline triggers are not supported in GitHub Actions")
	ErrUnsupportedUpdatePipeline  = newError("the update-pipeline feature is not supported, so you must always run 'halfpipe' to keep the workflow file up to date")
)

type Error struct {
	err   error
	file  string
	value string
}

func newError(msg string) Error {
	return Error{err: errors.New(msg)}
}

func (e Error) WithFile(file string) Error {
	return Error{err: e, file: file}
}

func (e Error) WithValue(value string) Error {
	return Error{err: e, value: value}
}

func (e Error) Error() string {
	s := e.err.Error()
	if e.value != "" {
		s = fmt.Sprintf("%s : %s", s, e.value)
	}
	if e.file != "" {
		s = fmt.Sprintf("%s (%s)", s, e.file)
	}
	return s
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}
	return t.err == e.err &&
		(t.file == "" || t.file == e.file) &&
		(t.value == "" || t.value == e.value)
}
