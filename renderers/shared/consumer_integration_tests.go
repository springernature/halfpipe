package shared

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"sort"
	"strings"
)

func ConsumerIntegrationTestScriptV2(vars manifest.Vars, cacheDirs []string) string {
	var envStrings []string
	for key := range vars {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)
	envOption := strings.Join(envStrings, " ")

	var cacheVolumeFlags []string
	for _, cacheVolume := range cacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cacheVolume, cacheVolume))
	}

	volumeOption := strings.Join(cacheVolumeFlags, " ")

	return fmt.Sprintf(`
export ENV_OPTIONS="%s"
export VOLUME_OPTIONS="%s"
run-cdc.sh`, envOption, volumeOption)
}

func ConsumerIntegrationTestScript(vars manifest.Vars, cacheDirs []string) string {
	var envStrings []string
	for key := range vars {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)
	envOption := strings.Join(envStrings, " ")

	var cacheVolumeFlags []string
	for _, cacheVolume := range cacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cacheVolume, cacheVolume))
	}

	volumeOption := strings.Join(cacheVolumeFlags, " ")

	return fmt.Sprintf(`# write git key to file
echo "${CONSUMER_GIT_KEY}" > .gitkey
chmod 600 .gitkey

set -ex

# get current revision of consumer, revert to HEAD if not found
REVISION=$(curl -fsSL "${CONSUMER_HOST}/internal/version" | jq -r '.revision' || echo "")
if [ "${REVISION}" == "" ]; then
  echo "Error fetching version of consumer from ${CONSUMER_HOST}/internal/version - using HEAD instead."
  REVISION=HEAD
fi

# clone consumer into /scratch/consumer. dir may already exist when concourse restarts task
rm -rf /scratch/consumer
GIT_SSH_COMMAND="ssh -o StrictHostKeychecking=no -i .gitkey" git clone ${GIT_CLONE_OPTIONS} ${CONSUMER_GIT_URI} /scratch/consumer
cd /scratch/consumer/${CONSUMER_PATH}

# checkout revision
git checkout ${REVISION}

# run the tests with docker-compose
# note: old system reads CF manifest env vars and sets them all here
docker-compose pull ${DOCKER_COMPOSE_SERVICE:-code}
docker-compose run --no-deps \
  --entrypoint "${CONSUMER_SCRIPT}" \
  -e ARTIFACTORY_PASSWORD \
  -e ARTIFACTORY_URL \
  -e ARTIFACTORY_USERNAME \
  -e DEPENDENCY_NAME=${PROVIDER_NAME} \
  -e ${PROVIDER_HOST_KEY}=${PROVIDER_HOST} \
  -e CDC_CONSUMER_NAME=${CONSUMER_NAME} \
  -e CDC_CONSUMER_VERSION=${REVISION} \
  -e CDC_PROVIDER_NAME=${PROVIDER_NAME} \
  -e CDC_PROVIDER_VERSION=${GIT_REVISION} \
  %s \
  %s \
  ${DOCKER_COMPOSE_SERVICE:-code}
`, envOption, volumeOption)
}
