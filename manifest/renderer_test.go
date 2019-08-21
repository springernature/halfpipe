package manifest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTheOnlyFieldThatShouldBeRenderedIsTypeIfAllTheOthersAreEmpty(t *testing.T) {
	man := Manifest{
		Pipeline: "yo",
		Triggers: TriggerList{
			GitTrigger{},
			TimerTrigger{},
			DockerTrigger{},
		},
		Tasks: TaskList{
			Run{},
			DockerCompose{},
			DeployCF{},
			DeployCF{
				PrePromote: TaskList{
					Run{},
				},
			},
			DockerPush{},
			ConsumerIntegrationTest{},
			DeployMLZip{},
			DeployMLModules{},
			Parallel{
				Tasks: TaskList{
					Run{},
					DeployCF{},
					DeployCF{
						PrePromote:TaskList{
							Run{},
						},
					},
				},
			},
		},
	}

	expected := `pipeline: yo
triggers:
- type: git
- type: timer
- type: docker
tasks:
- type: run
- type: docker-compose
- type: deploy-cf
- type: deploy-cf
  pre_promote:
  - type: run
- type: docker-push
- type: consumer-integration-test
- type: deploy-ml-zip
- type: deploy-ml-modules
- type: parallel
  tasks:
  - type: run
  - type: deploy-cf
  - type: deploy-cf
    pre_promote:
    - type: run
`

	yaml, err := Render(man)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(yaml))
}
