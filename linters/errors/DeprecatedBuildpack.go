package errors

type DeprecatedBuildpackError struct{}

func NewDeprecatedBuildpackError() DeprecatedBuildpackError {
	return DeprecatedBuildpackError{}
}

func (e DeprecatedBuildpackError) Error() string {
	return "Use of 'buildpack' attribute in manifest is deprecated in favor of 'buildpacks'. Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated for alternatives and other app manifest deprecations. This feature will be removed in the future."
}
