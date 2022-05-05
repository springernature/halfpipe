package linterrors

type DeprecatedDockerComposeVersionError struct{}

func (e DeprecatedDockerComposeVersionError) Error() string {
	return "the docker-compose file version used is deprecated. All services need to be under the 'services' key and 'Version' needs to be '2' or higher. Please see <https://docs.docker.com/compose/compose-file/compose-versioning/#versioning>"
}
