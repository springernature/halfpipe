package linterrors

import "fmt"

type DeprecatedBuildpackError struct{}

func NewDeprecatedBuildpackError() DeprecatedBuildpackError {
	return DeprecatedBuildpackError{}
}

func (e DeprecatedBuildpackError) Error() string {
	return "use of 'buildpack' attribute in manifest is deprecated in favor of 'buildpacks'. Please see <http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated>"
}

type UnversionedBuildpackError struct {
	buildpack string
}

func NewUnversionedBuildpackError(buildpack string) UnversionedBuildpackError {
	return UnversionedBuildpackError{buildpack}
}

func (e UnversionedBuildpackError) Error() string {
	return fmt.Sprintf("buildpack '%s' does not specify a version so the latest will be used on each deploy. It is recommended to pin to a version like this: %s#<VERSION>", e.buildpack, e.buildpack)
}

type MissingBuildpackError struct{}

func NewMissingBuildpackError() MissingBuildpackError {
	return MissingBuildpackError{}
}

func (e MissingBuildpackError) Error() string {
	return "no buildpack specified in manifest. Cloud Foundry will try to detect which system buildpack to use. Please see <https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#buildpack>"
}
