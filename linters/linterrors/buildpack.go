package linterrors

import "fmt"

type DeprecatedBuildpackError struct{}

func NewDeprecatedBuildpackError() DeprecatedBuildpackError {
	return DeprecatedBuildpackError{}
}

func (e DeprecatedBuildpackError) Error() string {
	return "Use of 'buildpack' attribute in manifest is deprecated in favor of 'buildpacks'. Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated for alternatives and other app manifest deprecations. This feature will be removed in the future."
}

type UnversionedBuildpackError struct {
	buildpack string
}

func NewUnversionedBuildpackError(buildpack string) UnversionedBuildpackError {
	return UnversionedBuildpackError{buildpack}
}

func (e UnversionedBuildpackError) Error() string {
	return fmt.Sprintf("Buildpack '%s' does not specify a version so the latest will be used on each deploy. It is recommended to pin to a version like this: %s#<VERSION>", e.buildpack, e.buildpack)
}
