package pipeline

import (
	"fmt"
	"strings"

	"sort"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (p pipeline) consumerIntegrationTestJob(task manifest.ConsumerIntegrationTest, man manifest.Manifest) *atc.JobConfig {

	consumerGitParts := strings.Split(task.Consumer, "/")
	consumerGitURI := fmt.Sprintf("git@github.com:springernature/%s", consumerGitParts[0])
	consumerGitPath := ""
	if len(consumerGitParts) > 1 {
		consumerGitPath = consumerGitParts[1]
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(man.Pipeline))

	// it is really just a special run job, so let's reuse that
	runTask := manifest.Run{
		Retries: task.Retries,
		Name:     task.Name,
		Script:   consumerIntegrationTestScript(task.Vars),
		Docker: manifest.Docker{
			Image:    config.DockerComposeImage,
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":       consumerGitURI,
			"CONSUMER_PATH":          consumerGitPath,
			"CONSUMER_SCRIPT":        task.Script,
			"CONSUMER_GIT_KEY":       "((github.private_key))",
			"CONSUMER_HOST":          task.ConsumerHost,
			"PROVIDER_NAME":          man.Pipeline,
			"PROVIDER_HOST_KEY":      providerHostKey,
			"PROVIDER_HOST":          task.ProviderHost,
			"DOCKER_COMPOSE_SERVICE": task.DockerComposeService,
			"GCR_PRIVATE_KEY":        "((gcr.private_key))",
		},
	}

	for key, val := range task.Vars {
		runTask.Vars[key] = val
	}

	job := p.runJob(runTask, man, true)
	return job
}

func consumerIntegrationTestScript(vars manifest.Vars) string {
	var envStrings []string
	for key := range vars {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)

	return fmt.Sprintf(`\docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io

# write git key to file
echo "${CONSUMER_GIT_KEY}" > .gitkey
chmod 600 .gitkey

set -x

# get current revision of consumer
REVISION=$(curl "${CONSUMER_HOST}/internal/version" | jq -r '.revision')

# clone consumer into "consumer-repo" dir
GIT_SSH_COMMAND="ssh -o StrictHostKeychecking=no -i .gitkey" git clone ${CONSUMER_GIT_URI} consumer-repo
cd consumer-repo/${CONSUMER_PATH}

# checkout revision
git checkout ${REVISION}

# run the tests with docker-compose
# note: old system reads CF manifest env vars and sets them all here
docker-compose run --no-deps \
  --entrypoint "${CONSUMER_SCRIPT}" \
  -e DEPENDENCY_NAME=${PROVIDER_NAME} \
  -e ${PROVIDER_HOST_KEY}=${PROVIDER_HOST} %s \
  ${DOCKER_COMPOSE_SERVICE:-code}
`, strings.Join(envStrings, " "))
}
