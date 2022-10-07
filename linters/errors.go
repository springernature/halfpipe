package linters

import (
	"errors"
	"fmt"
)

var (
	ErrMissingField    = NewError("required field missing")
	NewErrMissingField = func(field string) Error { return ErrMissingField.WithValue(field) }

	ErrInvalidField    = NewError("invalid field")
	NewErrInvalidField = func(field string, reason string) Error { return ErrInvalidField.WithValue(field).WithValue(reason) }

	ErrFileNotFound      = NewError("file not found")
	ErrFileCannotRead    = NewError("file cannot be read")
	ErrFileNotAFile      = NewError("not a file")
	ErrFileEmpty         = NewError("file is empty")
	ErrFileNotExecutable = NewError("file is not executable")
	ErrFileInvalid       = NewError("file is invalid")

	ErrCFMissingRoutes        = NewError("cf application must have at least one route")
	ErrCFMissingName          = NewError("cf application missing 'name'")
	ErrCFRoutesAndNoRoute     = NewError("cf application cannot have both 'routes' and 'no-route'")
	ErrCFNoRouteHealthcheck   = NewError("cf application with 'no-route: true' requires 'health-check-type: process'")
	ErrCFRouteScheme          = NewError("cf application route must not start with http(s)://")
	ErrCFRouteMissing         = NewError("cf application routes must contain sso_route")
	ErrCFMultipleApps         = NewError("cf manifest must have exactly 1 application")
	ErrCFBuildpackUnversioned = NewError("buildpack specified without version so the latest will be used on each deploy")
	ErrCFBuildpackMissing     = NewError("buildpack missing. Cloud Foundry will try to detect which system buildpack to use. Please see <https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#buildpack>")
	ErrCFBuildpackDeprecated  = NewError("'buildpack' is deprecated in favour of 'buildpacks'. Please see <http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated>")
	ErrCFArtifactAndDocker    = NewError("cannot combine 'deploy_artifact' in the halfpipe task and 'docker' in the cf manifest")
	ErrCFFromArtifact         = NewError("this file must be saved as an artifact in a previous task")
	ErrCFPrePromoteArtifact   = NewError("cannot have pre promote tasks with CF manifest restored from artifact")

	ErrUnsupportedRegistry  = NewError("image must be from halfpipe registry. Please see <https://ee.public.springernature.app/rel-eng/docker-registry/>")
	ErrDockerPushTag        = NewError("the field 'tag' is no longer used and is safe to delete")
	ErrDockerComposeVersion = NewError("the docker-compose file version used is deprecated. All services must be under the 'services' key and 'Version' must be '2' or higher. Please see <https://docs.docker.com/compose/compose-file/compose-versioning/#versioning>")
	ErrMultipleTriggers     = NewError("cannot have multiple triggers of this type")

	ErrVelaVariableMissing = NewError("vela manifest variable is not specified in halfpipe manifest")

	ErrUnsupportedManualTrigger   = NewError("manual_trigger on individual tasks is not supported in GitHub Actions. It is supported at the workflow level in git trigger options")
	ErrUnsupportedRolling         = NewError("cf rolling deploys are not supported in GitHub Actions")
	ErrDockerTriggerLoop          = NewError("cannot push docker image that is also a trigger as it will create a loop")
	ErrUnsupportedCovenant        = NewError("covenant is not supported in GitHub Actions")
	ErrUnsupportedGitPrivateKey   = NewError("git private_key is not supported in GitHub Actions")
	ErrUnsupportedGitUri          = NewError("git uri is not supported in GitHub Actions")
	ErrUnsupportedPipelineTrigger = NewError("pipeline triggers are not supported in GitHub Actions")
	ErrUnsupportedUpdatePipeline  = NewError("the update-pipeline feature is not supported, so you must always run 'halfpipe' to keep the workflow file up to date")
)

type Error struct {
	err   error
	file  string
	value string
}

func NewError(msg string) Error {
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
