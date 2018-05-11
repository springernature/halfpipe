package pipeline

import (
	"fmt"
	"strings"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (p pipeline) consumerIntegrationTestJob(task manifest.ConsumerIntegrationTest, man manifest.Manifest, testRoute string) *atc.JobConfig {

	consumerGitParts := strings.Split(task.Consumer, "/")
	consumerGitURI := fmt.Sprintf("git@github.com:springernature/%s", consumerGitParts[0])
	consumerGitPath := ""
	if len(consumerGitParts) > 1 {
		consumerGitPath = consumerGitParts[1]
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(man.Pipeline))

	// it is really just a special run job, so let's reuse that
	runTask := manifest.Run{
		Name:   task.Name,
		Script: consumerIntegrationTestScript,
		Docker: manifest.Docker{
			Image:    config.ConsumerIntegrationTestImage,
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":  consumerGitURI,
			"CONSUMER_PATH":     consumerGitPath,
			"CONSUMER_SCRIPT":   task.Script,
			"CONSUMER_GIT_KEY":  "((github.private_key))",
			"CONSUMER_HOST":     task.Host,
			"PROVIDER_NAME":     man.Pipeline,
			"PROVIDER_HOST_KEY": providerHostKey,
			"PROVIDER_HOST":     testRoute,
		},
	}
	job := p.runJob(runTask, man)
	job.Plan[0].Privileged = true
	return job
}

const consumerIntegrationTestScript = `\source /docker-lib.sh
start_docker

# get current revision of consumer
REVISION=$(curl "${CONSUMER_HOST}/internal/version" | jq -r '.revision')

# write git key to file
echo "${CONSUMER_GIT_KEY}" > .gitkey
chmod 600 .gitkey

set -x

# clone consumer into "consumer-repo" dir
GIT_SSH_COMMAND="ssh -o StrictHostKeychecking=no -i .gitkey" git clone ${CONSUMER_GIT_URI} consumer-repo
cd consumer-repo/${CONSUMER_PATH}

# checkout revision
git checkout ${REVISION}

# run the tests in code container
# note: old system reads CF manifest env vars and sets them all here
docker-compose run --no-deps \
  -e DEPENDENCY_NAME=${PROVIDER_NAME} \
  -e ${PROVIDER_HOST_KEY}=${PROVIDER_HOST} \
  code \
  /home/dev/code/${CONSUMER_SCRIPT}

rc=$?
docker-compose down
[ $rc -eq 0 ] || exit $rc
`
