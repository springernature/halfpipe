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

	ErrDeprecatedField = newError("deprecated field")
	NewDeprecatedField = func(field string, reason string) Error { return ErrDeprecatedField.WithValue(field).WithValue(reason) }

	ErrFileNotFound      = newError("file not found")
	ErrFileCannotRead    = newError("file cannot be read")
	ErrFileNotAFile      = newError("not a file")
	ErrFileEmpty         = newError("file is empty")
	ErrFileNotExecutable = newError("file is not executable")
	ErrFileInvalid       = newError("file is invalid")

	ErrCFMissingRoutes         = newError("cf application must have at least one route")
	ErrCFMissingName           = newError("cf application missing 'name'")
	ErrCFRoutesAndNoRoute      = newError("cf application cannot have both 'routes' and 'no-route'")
	ErrCFNoRouteHealthcheck    = newError("cf application with 'no-route: true' requires 'health-check-type: process'")
	ErrCFRouteScheme           = newError("cf application route must not start with http(s)://")
	ErrCFRouteMissing          = newError("cf application routes must contain sso_route")
	ErrCFMultipleApps          = newError("cf manifest must have exactly 1 application")
	ErrCFBuildpackUnversioned  = newError("buildpack specified without version so the latest will be used on each deploy")
	ErrCFBuildpackMissing      = newError("buildpack missing. Cloud Foundry will try to detect which system buildpack to use. Please see <https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#buildpack>")
	ErrCFBuildpackDeprecated   = newError("'buildpack' is deprecated in favour of 'buildpacks'. Please see <http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated>")
	ErrCFArtifactAndDocker     = newError("cannot combine 'deploy_artifact' in the halfpipe task and 'docker' in the cf manifest")
	ErrCFFromArtifact          = newError("this file must be saved as an artifact in a previous task")
	ErrCFPrePromoteArtifact    = newError("cannot have pre promote tasks with CF manifest restored from artifact")
	ErrCFCandidateRouteTooLong = newError("cf does not allow routes of more than 64 characters")

	ErrCFLabelTeamWillBeOverwritten = newError("deployment will overwrite metadata.labels.team that was set in the CF manifest").AsWarning()
	ErrCFLabelProductIsMissing      = newError("CF manifest is missing 'product' label. If 'product' is set on the CF space you can safely ignore this warning.").AsWarning()
	ErrCFLabelEnvironmentIsMissing  = newError("CF manifest is missing 'environment' label. If 'environment' is set on the CF space you can safely ignore this warning.").AsWarning()

	ErrUnsupportedRegistry = newError("image must be from halfpipe registry. Please see <https://ee.public.springernature.app/rel-eng/docker-registry/>")
	ErrDockerPushTag       = newError("the field 'tag' is no longer used and is safe to delete")

	ErrDockerPlatformUnknown = newError("only linux/amd64 and/or linux/arm64 are supported")
	ErrDockerComposeVersion  = newError("the docker-compose file version used is deprecated. All services must be under the 'services' key and 'Version' must be '2' or higher. Please see <https://docs.docker.com/compose/compose-file/compose-versioning/#versioning>")
	ErrDockerVarSecret       = newError("using a secret in docker build vars is not secure. See the 'secrets' option of the docker-push task")
	ErrDockerRegistry        = newError("image should be pushed to 'eu.gcr.io/halfpipe-io/<team>/<image>'")

	ErrMultipleTriggers = newError("cannot have multiple triggers of this type")

	ErrVelaVariableMissing        = newError("vela manifest variable is not specified in halfpipe manifest")
	ErrVelaNamespace              = newError("vela namespace must start with 'katee-'")
	ErrVelaDeploymentCheckTimeout = newError("deployment_check_timeout is deprecated. Please use max_checks and check_interval")
	ErrVelaEnvironment            = newError("the field 'environment' is no longer used and is safe to delete")

	ErrUnsupportedManualTrigger   = newError("manual_trigger on individual tasks is not supported in GitHub Actions. It is supported at the workflow level in git trigger options")
	ErrUnsupportedRolling         = newError("cf rolling deploys are not supported in GitHub Actions")
	ErrDockerTriggerLoop          = newError("cannot push docker image that is also a trigger as it will create a loop")
	ErrUnsupportedGitPrivateKey   = newError("git private_key is not supported in GitHub Actions")
	ErrUnsupportedGitUri          = newError("git uri is not supported in GitHub Actions")
	ErrUnsupportedPipelineTrigger = newError("pipeline triggers are not supported in GitHub Actions")

	ErrSlackSuccessMessageFieldDeprecated = newError("'slack_success_message' is deprecated, please use new notification structure")
	ErrSlackFailureMessageFieldDeprecated = newError("'slack_failure_message' is deprecated, please use new notification structure")
	ErrOnlySlackOrTeamsAllowed            = newError("You cannot define both 'slack' and 'teams' in the notifications")
)

type Error struct {
	err   error
	value string
	level string
}

func newError(msg string) Error {
	return Error{err: errors.New(msg), level: "error"}
}

func (e Error) WithFile(file string) Error {
	return Error{err: e, level: e.level, value: " (" + file + ")"}
}

func (e Error) WithValue(value string) Error {
	return Error{err: e, level: e.level, value: ": " + value}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s%s", e.err.Error(), e.value)
}

func (e Error) AsWarning() Error {
	return Error{err: e.err, level: "warning", value: e.value}
}

func (e Error) IsWarning() bool {
	return e.level == "warning"
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}
	return t.err == e.err && (t.value == "" || t.value == e.value)
}
